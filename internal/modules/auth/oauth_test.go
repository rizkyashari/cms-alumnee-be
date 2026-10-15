package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

type fakeGoogleVerifier struct {
	wantClientID string
	wantIDToken  string
	wantNonce    string
	user         providerUser
	err          error
}

func (verifier fakeGoogleVerifier) verify(_ context.Context, _ *http.Client, clientID, idToken, nonce string) (providerUser, error) {
	if clientID != verifier.wantClientID || idToken != verifier.wantIDToken || nonce != verifier.wantNonce {
		return providerUser{}, errors.New("unexpected Google OIDC verifier input")
	}
	return verifier.user, verifier.err
}

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func oauthTestService() *Service {
	service := NewService(nil, false, "")
	service.ConfigureIntegrations(Integrations{
		PublicBaseURL:    "https://alumni.example",
		TurnstileSiteKey: "public-site-key",
		TurnstileSecret:  "private-turnstile-secret",
		GoogleClientID:   "google-client-id",
		GoogleSecret:     "private-google-secret",
		FacebookClientID: "facebook-client-id",
		FacebookSecret:   "private-facebook-secret",
	})
	return service
}

func TestOAuthProviderDiscoveryOnlyExposesReadyProvidersAndPublicKey(t *testing.T) {
	tests := []struct {
		name          string
		integrations  Integrations
		wantProviders []string
		wantSiteKey   string
	}{
		{
			name: "both providers and captcha",
			integrations: Integrations{
				TurnstileSiteKey: "public-site-key", TurnstileSecret: "private-turnstile-secret",
				GoogleClientID: "google-client-id", GoogleSecret: "private-google-secret",
				FacebookClientID: "facebook-client-id", FacebookSecret: "private-facebook-secret",
			},
			wantProviders: []string{"google", "facebook"},
			wantSiteKey:   "public-site-key",
		},
		{
			name: "google only",
			integrations: Integrations{
				TurnstileSiteKey: "public-site-key", TurnstileSecret: "private-turnstile-secret",
				GoogleClientID: "google-client-id", GoogleSecret: "private-google-secret",
			},
			wantProviders: []string{"google"},
			wantSiteKey:   "public-site-key",
		},
		{
			name: "oauth disabled without captcha",
			integrations: Integrations{
				GoogleClientID: "google-client-id", GoogleSecret: "private-google-secret",
			},
			wantProviders: []string{},
		},
		{
			name: "partial provider credentials",
			integrations: Integrations{
				TurnstileSiteKey: "public-site-key", TurnstileSecret: "private-turnstile-secret",
				GoogleClientID: "google-client-id",
			},
			wantProviders: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(nil, false, "")
			service.ConfigureIntegrations(test.integrations)
			handler := NewHandler(service, nil)
			router := chi.NewRouter()
			handler.RegisterRoutes(router)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/providers", nil))

			if recorder.Code != http.StatusOK || recorder.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("unexpected response status/cache headers: status=%d headers=%v", recorder.Code, recorder.Header())
			}
			responseBody := recorder.Body.String()
			var response struct {
				Content struct {
					Providers      []string `json:"providers"`
					CaptchaSiteKey string   `json:"captcha_site_key"`
				} `json:"content"`
			}
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode provider configuration: %v", err)
			}
			if strings.Join(response.Content.Providers, ",") != strings.Join(test.wantProviders, ",") {
				t.Fatalf("providers = %v, want %v", response.Content.Providers, test.wantProviders)
			}
			if response.Content.CaptchaSiteKey != test.wantSiteKey {
				t.Fatalf("captcha site key = %q, want %q", response.Content.CaptchaSiteKey, test.wantSiteKey)
			}
			for _, secret := range []string{"private-google-secret", "private-facebook-secret", "private-turnstile-secret"} {
				if strings.Contains(responseBody, secret) {
					t.Fatalf("provider config leaked a secret")
				}
			}
		})
	}
}

func TestOAuthStartRequiresServerVerifiedCaptchaAndSetsStateCookie(t *testing.T) {
	service := oauthTestService()
	service.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != turnstileVerifyURL {
			t.Fatalf("unexpected external URL: %s", request.URL)
		}
		if err := request.ParseForm(); err != nil {
			t.Fatalf("parse Turnstile request: %v", err)
		}
		if request.Form.Get("secret") != "private-turnstile-secret" || request.Form.Get("response") != "captcha-token" {
			t.Fatalf("unexpected Turnstile request form")
		}
		return jsonResponse(http.StatusOK, `{"success":true,"hostname":"alumni.example"}`), nil
	})}
	handler := NewHandler(service, nil)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/google/start", strings.NewReader(`{"captcha_token":"captcha-token"}`))
	request.Header.Set("Content-Type", "application/json")
	request.RemoteAddr = "127.0.0.1:1234"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("OAuth start status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Content struct {
			AuthorizationURL string `json:"authorization_url"`
		} `json:"content"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode authorization response: %v", err)
	}
	authorizationURL, err := url.Parse(response.Content.AuthorizationURL)
	if err != nil {
		t.Fatalf("parse authorization URL: %v", err)
	}
	query := authorizationURL.Query()
	if query.Get("client_id") != "google-client-id" ||
		query.Get("redirect_uri") != "https://alumni.example/api/v1/auth/oauth/google/callback" ||
		query.Get("response_type") != "code" || query.Get("state") == "" || query.Get("nonce") == "" {
		t.Fatalf("authorization URL parameters are incomplete: %s", authorizationURL)
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("state/nonce cookies = %d, want two", len(cookies))
	}
	cookiesByName := make(map[string]*http.Cookie, len(cookies))
	for _, cookie := range cookies {
		cookiesByName[cookie.Name] = cookie
	}
	stateCookie := cookiesByName["alumnihub_oauth_state_google"]
	if stateCookie == nil || stateCookie.Name != "alumnihub_oauth_state_google" ||
		stateCookie.Value != query.Get("state") ||
		!stateCookie.HttpOnly || !stateCookie.Secure ||
		stateCookie.SameSite != http.SameSiteLaxMode ||
		stateCookie.Path != "/api/v1/auth/oauth/google" {
		t.Fatalf("unexpected OAuth state cookie: %+v", stateCookie)
	}
	nonceCookie := cookiesByName["alumnihub_oauth_nonce_google"]
	if nonceCookie == nil || nonceCookie.Value != query.Get("nonce") ||
		!nonceCookie.HttpOnly || !nonceCookie.Secure || nonceCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected Google nonce cookie: %+v", nonceCookie)
	}
}

func TestOAuthStartRejectsCaptchaFailureAndProviderMismatch(t *testing.T) {
	tests := []struct {
		name            string
		captchaResponse string
		provider        string
		wantStatus      int
	}{
		{name: "rejected captcha", captchaResponse: `{"success":false}`, provider: "google", wantStatus: http.StatusUnauthorized},
		{name: "wrong captcha hostname", captchaResponse: `{"success":true,"hostname":"attacker.example"}`, provider: "google", wantStatus: http.StatusUnauthorized},
		{name: "unsupported provider", captchaResponse: `{"success":true,"hostname":"alumni.example"}`, provider: "unknown", wantStatus: http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := oauthTestService()
			service.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return jsonResponse(http.StatusOK, test.captchaResponse), nil
			})}
			handler := NewHandler(service, nil)
			router := chi.NewRouter()
			handler.RegisterRoutes(router)
			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/auth/oauth/"+test.provider+"/start",
				strings.NewReader(`{"captcha_token":"captcha-token"}`),
			)
			request.Header.Set("Content-Type", "application/json")
			request.RemoteAddr = "127.0.0.1:1234"
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Fatalf("OAuth start status = %d, want %d; body=%s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			if len(recorder.Result().Cookies()) > 0 {
				t.Fatal("failed OAuth start unexpectedly issued a state cookie")
			}
		})
	}
}

func TestFacebookOAuthStartUsesConfiguredFacebookAuthorization(t *testing.T) {
	service := oauthTestService()
	service.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{"success":true,"hostname":"alumni.example"}`), nil
	})}
	handler := NewHandler(service, nil)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/facebook/start", strings.NewReader(`{"captcha_token":"captcha-token"}`))
	request.Header.Set("Content-Type", "application/json")
	request.RemoteAddr = "127.0.0.1:1234"
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Facebook OAuth start status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Content struct {
			AuthorizationURL string `json:"authorization_url"`
		} `json:"content"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode Facebook authorization response: %v", err)
	}
	authorizationURL, err := url.Parse(response.Content.AuthorizationURL)
	if err != nil {
		t.Fatalf("parse Facebook authorization URL: %v", err)
	}
	query := authorizationURL.Query()
	if authorizationURL.Host != "www.facebook.com" ||
		query.Get("client_id") != "facebook-client-id" ||
		query.Get("scope") != "email,public_profile" ||
		query.Get("redirect_uri") != "https://alumni.example/api/v1/auth/oauth/facebook/callback" ||
		query.Get("state") == "" {
		t.Fatalf("Facebook authorization URL is incomplete: %s", authorizationURL)
	}
	if query.Get("nonce") != "" || len(recorder.Result().Cookies()) != 1 {
		t.Fatal("Facebook OAuth unexpectedly used Google OIDC nonce")
	}
}

func TestOAuthCallbackRedirectsSafeErrorsToLogin(t *testing.T) {
	tests := []struct {
		name            string
		state           string
		query           string
		includeNonce    bool
		providerFailure bool
		wantCode        string
	}{
		{name: "invalid state", state: "cookie-state", query: "state=attacker-state&code=ignored", wantCode: "oauth_state_invalid"},
		{name: "missing Google nonce", state: "valid-state", query: "state=valid-state&code=ignored", wantCode: "oauth_state_invalid"},
		{name: "provider cancellation", state: "valid-state", query: "state=valid-state&error=access_denied&error_description=private+provider+details", includeNonce: true, wantCode: "oauth_cancelled"},
		{name: "provider token exchange failure", state: "valid-state", query: "state=valid-state&code=authorization-code", includeNonce: true, providerFailure: true, wantCode: "oauth_provider_failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := oauthTestService()
			if test.providerFailure {
				service.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					return jsonResponse(http.StatusBadGateway, `{"error":"provider unavailable"}`), nil
				})}
			}
			handler := NewHandler(service, nil)
			router := chi.NewRouter()
			handler.RegisterRoutes(router)
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/auth/oauth/google/callback?"+test.query,
				nil,
			)
			request.AddCookie(&http.Cookie{Name: "alumnihub_oauth_state_google", Value: test.state})
			if test.includeNonce {
				request.AddCookie(&http.Cookie{Name: "alumnihub_oauth_nonce_google", Value: "valid-nonce"})
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusSeeOther {
				t.Fatalf("callback status = %d, want %d; body=%s", recorder.Code, http.StatusSeeOther, recorder.Body.String())
			}
			location, err := url.Parse(recorder.Header().Get("Location"))
			if err != nil {
				t.Fatalf("parse callback redirect: %v", err)
			}
			if location.String() != "https://alumni.example/login/?oauth_error="+test.wantCode {
				t.Fatalf("callback redirect = %s, want safe login error %s", location, test.wantCode)
			}
			if strings.Contains(location.String(), "private") || strings.Contains(location.String(), "attacker-state") {
				t.Fatalf("callback redirect leaked provider/state details: %s", location)
			}
		})
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestOAuthCaptchaReturnsUnavailableOnProviderFailure(t *testing.T) {
	service := oauthTestService()
	service.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("provider unavailable")
	})}
	err := service.verifyCaptcha(context.Background(), "captcha-token", "127.0.0.1")
	if !errors.Is(err, ErrCaptchaUnavailable) {
		t.Fatalf("verifyCaptcha error = %v, want ErrCaptchaUnavailable", err)
	}
}

func TestOAuthCodeExchangeLoadsProviderIdentity(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		userInfoJSON string
		wantSubject  string
		wantEmail    string
		wantVerified bool
	}{
		{
			name:         "google verified email",
			providerName: "google",
			userInfoJSON: `{"sub":"google-subject","email":"alumni@example.com","email_verified":true,"name":"Alumni Name"}`,
			wantSubject:  "google-subject", wantEmail: "alumni@example.com", wantVerified: true,
		},
		{
			name:         "facebook email is not trusted as verified",
			providerName: "facebook",
			userInfoJSON: `{"id":"facebook-subject","email":"alumni@example.com","name":"Alumni Name"}`,
			wantSubject:  "facebook-subject", wantEmail: "alumni@example.com", wantVerified: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/token":
					if err := r.ParseForm(); err != nil {
						t.Fatalf("parse token request: %v", err)
					}
					if r.Form.Get("client_secret") != "provider-secret" || r.Form.Get("code") != "authorization-code" {
						t.Fatalf("unexpected provider token request")
					}
					if test.providerName == "google" {
						_, _ = io.WriteString(w, `{"access_token":"provider-access-token","id_token":"signed-google-id-token"}`)
					} else {
						_, _ = io.WriteString(w, `{"access_token":"provider-access-token"}`)
					}
				case "/userinfo":
					if r.Header.Get("Authorization") != "Bearer provider-access-token" {
						t.Fatalf("provider access token was not sent as a bearer token")
					}
					_, _ = io.WriteString(w, test.userInfoJSON)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			provider := oauthProvider{
				clientID:     "provider-client",
				clientSecret: "provider-secret",
				tokenURL:     server.URL + "/token",
				userInfoURL:  server.URL + "/userinfo",
				redirectURI:  "https://alumni.example/api/v1/auth/oauth/" + test.providerName + "/callback",
			}
			googleVerifier := fakeGoogleVerifier{
				wantClientID: "provider-client", wantIDToken: "signed-google-id-token", wantNonce: "expected-nonce",
				user: providerUser{
					sub: "google-subject", email: "alumni@example.com",
					emailVerified: true, name: "Alumni Name",
				},
			}
			user, err := exchangeOAuthCode(context.Background(), server.Client(), googleVerifier, test.providerName, provider, "authorization-code", "expected-nonce")
			if err != nil {
				t.Fatalf("exchangeOAuthCode() error = %v", err)
			}
			if user.sub != test.wantSubject || user.email != test.wantEmail || user.name != "Alumni Name" || user.emailVerified != test.wantVerified {
				t.Fatalf("provider identity = %+v", user)
			}
		})
	}
}

func TestGoogleNonceComparisonRequiresExactMatch(t *testing.T) {
	if !constantTimeEqual("expected-nonce", "expected-nonce") {
		t.Fatal("matching nonce was rejected")
	}
	for _, nonce := range []string{"", "other-nonce", "expected-nonce-longer"} {
		if constantTimeEqual("expected-nonce", nonce) {
			t.Fatalf("mismatched nonce %q was accepted", nonce)
		}
	}
}

func TestGoogleOIDCVerifierChecksSignatureIssuerAudienceExpiryAndNonce(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate test signing key: %v", err)
	}
	tests := []struct {
		name       string
		claims     map[string]any
		corruptSig bool
		wantErr    bool
	}{
		{
			name:   "valid id token",
			claims: googleIDTokenClaims("expected-client", "expected-nonce", time.Now().Add(time.Hour).Unix()),
		},
		{
			name: "invalid issuer",
			claims: func() map[string]any {
				claims := googleIDTokenClaims("expected-client", "expected-nonce", time.Now().Add(time.Hour).Unix())
				claims["iss"] = "https://attacker.example"
				return claims
			}(),
			wantErr: true,
		},
		{
			name:    "invalid audience",
			claims:  googleIDTokenClaims("other-client", "expected-nonce", time.Now().Add(time.Hour).Unix()),
			wantErr: true,
		},
		{
			name:    "expired token",
			claims:  googleIDTokenClaims("expected-client", "expected-nonce", time.Now().Add(-time.Hour).Unix()),
			wantErr: true,
		},
		{
			name:    "nonce mismatch",
			claims:  googleIDTokenClaims("expected-client", "attacker-nonce", time.Now().Add(time.Hour).Unix()),
			wantErr: true,
		},
		{
			name:       "invalid signature",
			claims:     googleIDTokenClaims("expected-client", "expected-nonce", time.Now().Add(time.Hour).Unix()),
			corruptSig: true,
			wantErr:    true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rawToken := signedGoogleIDToken(t, privateKey, test.claims)
			if test.corruptSig {
				parts := strings.Split(rawToken, ".")
				signature, err := base64.RawURLEncoding.DecodeString(parts[2])
				if err != nil {
					t.Fatalf("decode test signature: %v", err)
				}
				signature[0] ^= 0x01
				parts[2] = base64.RawURLEncoding.EncodeToString(signature)
				rawToken = strings.Join(parts, ".")
			}
			client := googleOIDCTestClient(privateKey)
			user, err := (googleOIDCVerifier{}).verify(context.Background(), client, "expected-client", rawToken, "expected-nonce")
			if test.wantErr {
				if err == nil {
					t.Fatal("invalid Google ID token was accepted")
				}
				return
			}
			if err != nil {
				t.Fatalf("valid Google ID token rejected: %v", err)
			}
			if user.sub != "google-subject" || user.email != "alumni@example.com" || !user.emailVerified || user.name != "Alumni Name" {
				t.Fatalf("unexpected verified Google claims: %+v", user)
			}
		})
	}
}

func googleIDTokenClaims(audience, nonce string, expiresAt int64) map[string]any {
	return map[string]any{
		"iss":            "https://accounts.google.com",
		"sub":            "google-subject",
		"aud":            audience,
		"exp":            expiresAt,
		"iat":            time.Now().Unix(),
		"nonce":          nonce,
		"email":          "alumni@example.com",
		"email_verified": true,
		"name":           "Alumni Name",
	}
}

func signedGoogleIDToken(t *testing.T, key *rsa.PrivateKey, claims map[string]any) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "RS256", "kid": "test-google-key", "typ": "JWT"})
	if err != nil {
		t.Fatalf("encode test JWT header: %v", err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("encode test JWT claims: %v", err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign test JWT: %v", err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func googleOIDCTestClient(key *rsa.PrivateKey) *http.Client {
	exponent := big.NewInt(int64(key.PublicKey.E)).Bytes()
	jwks := map[string]any{
		"keys": []map[string]string{{
			"kty": "RSA",
			"kid": "test-google-key",
			"alg": "RS256",
			"use": "sig",
			"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(exponent),
		}},
	}
	discovery := `{"issuer":"https://accounts.google.com","jwks_uri":"https://keys.example/google-jwks","id_token_signing_alg_values_supported":["RS256"]}`
	return &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.String() {
		case "https://accounts.google.com/.well-known/openid-configuration":
			return jsonResponse(http.StatusOK, discovery), nil
		case "https://keys.example/google-jwks":
			body, err := json.Marshal(jwks)
			if err != nil {
				return nil, err
			}
			return jsonResponse(http.StatusOK, string(body)), nil
		default:
			return nil, errors.New("unexpected OIDC verifier request")
		}
	})}
}
