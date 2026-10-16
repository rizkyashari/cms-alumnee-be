package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type oauthProvider struct {
	clientID     string
	clientSecret string
	authURL      string
	tokenURL     string
	userInfoURL  string
	redirectURI  string
}

type providerUser struct {
	sub           string
	email         string
	emailVerified bool
	name          string
}

type googleIDTokenVerifier interface {
	verify(context.Context, *http.Client, string, string, string) (providerUser, error)
}

type googleOIDCVerifier struct{}

func (googleOIDCVerifier) verify(ctx context.Context, client *http.Client, clientID, rawIDToken, expectedNonce string) (providerUser, error) {
	ctx = oidc.ClientContext(ctx, client)
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return providerUser{}, fmt.Errorf("discover google oidc provider: %w", err)
	}
	idToken, err := provider.VerifierContext(ctx, &oidc.Config{ClientID: clientID}).Verify(ctx, rawIDToken)
	if err != nil {
		return providerUser{}, fmt.Errorf("verify google id token: %w", err)
	}
	var claims struct {
		Nonce         string `json:"nonce"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return providerUser{}, fmt.Errorf("read google id token claims: %w", err)
	}
	if expectedNonce == "" || !constantTimeEqual(claims.Nonce, expectedNonce) {
		return providerUser{}, errors.New("google id token nonce does not match")
	}
	return providerUser{
		sub: idToken.Subject, email: claims.Email,
		emailVerified: claims.EmailVerified, name: claims.Name,
	}, nil
}

func (h *Handler) registerOAuthRoutes(router chi.Router) {
	router.Get("/api/v1/auth/oauth/providers", h.oauthProviders)
	router.Post("/api/v1/auth/oauth/{provider}/start", h.startOAuth)
	router.Get("/api/v1/auth/oauth/{provider}/callback", h.oauthCallback)
}

func (s *Service) provider(name string) (oauthProvider, bool) {
	redirectURI := strings.TrimRight(s.integrations.PublicBaseURL, "/") + "/api/v1/auth/oauth/" + name + "/callback"
	switch name {
	case "google":
		if s.integrations.GoogleClientID == "" || s.integrations.GoogleSecret == "" {
			return oauthProvider{}, false
		}
		return oauthProvider{
			clientID: s.integrations.GoogleClientID, clientSecret: s.integrations.GoogleSecret,
			authURL:     "https://accounts.google.com/o/oauth2/v2/auth",
			tokenURL:    "https://oauth2.googleapis.com/token",
			userInfoURL: "https://openidconnect.googleapis.com/v1/userinfo",
			redirectURI: redirectURI,
		}, true
	case "facebook":
		if s.integrations.FacebookClientID == "" || s.integrations.FacebookSecret == "" {
			return oauthProvider{}, false
		}
		return oauthProvider{
			clientID: s.integrations.FacebookClientID, clientSecret: s.integrations.FacebookSecret,
			authURL:     "https://www.facebook.com/v21.0/dialog/oauth",
			tokenURL:    "https://graph.facebook.com/v21.0/oauth/access_token",
			userInfoURL: "https://graph.facebook.com/me?fields=id,name,email",
			redirectURI: redirectURI,
		}, true
	default:
		return oauthProvider{}, false
	}
}

func (h *Handler) oauthProviders(w http.ResponseWriter, _ *http.Request) {
	providers := make([]string, 0, 2)
	captchaConfigured := h.service.integrations.TurnstileSiteKey != "" && h.service.integrations.TurnstileSecret != ""
	if captchaConfigured {
		if _, enabled := h.service.provider("google"); enabled {
			providers = append(providers, "google")
		}
		if _, enabled := h.service.provider("facebook"); enabled {
			providers = append(providers, "facebook")
		}
	}
	captchaSiteKey := ""
	if len(providers) > 0 {
		captchaSiteKey = h.service.integrations.TurnstileSiteKey
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "ok",
		"content": map[string]any{
			"providers":              providers,
			"captcha_site_key":       captchaSiteKey,
			"password_login_enabled": h.service.passwordLoginEnabled,
		},
	})
}

func (h *Handler) startOAuth(w http.ResponseWriter, r *http.Request) {
	if !h.limiter.Allow(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "too many login attempts")
		return
	}
	providerName := chi.URLParam(r, "provider")
	provider, enabled := h.service.provider(providerName)
	if !enabled {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	var input struct {
		CaptchaToken string `json:"captcha_token"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.service.verifyCaptcha(r.Context(), input.CaptchaToken, clientIP(r)); err != nil {
		if errors.Is(err, ErrCaptchaRejected) {
			writeError(w, http.StatusUnauthorized, "captcha verification failed")
			return
		}
		h.reportFailure("captcha_verification", err)
		writeError(w, http.StatusServiceUnavailable, "captcha verification unavailable")
		return
	}
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		writeError(w, http.StatusInternalServerError, "could not start login")
		return
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	nonce := ""
	if providerName == "google" {
		nonceBytes := make([]byte, 32)
		if _, err := rand.Read(nonceBytes); err != nil {
			writeError(w, http.StatusInternalServerError, "could not start login")
			return
		}
		nonce = base64.RawURLEncoding.EncodeToString(nonceBytes)
	}
	http.SetCookie(w, &http.Cookie{
		Name: "alumnihub_oauth_state_" + providerName, Value: state,
		Path: "/api/v1/auth/oauth/" + providerName, MaxAge: 300,
		HttpOnly: true, Secure: h.service.cookieSecure, SameSite: http.SameSiteLaxMode,
	})
	if providerName == "google" {
		http.SetCookie(w, &http.Cookie{
			Name: "alumnihub_oauth_nonce_google", Value: nonce,
			Path: "/api/v1/auth/oauth/google", MaxAge: 300,
			HttpOnly: true, Secure: h.service.cookieSecure, SameSite: http.SameSiteLaxMode,
		})
	}
	authorizationURL, err := url.Parse(provider.authURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not start login")
		return
	}
	query := authorizationURL.Query()
	query.Set("client_id", provider.clientID)
	query.Set("redirect_uri", provider.redirectURI)
	query.Set("response_type", "code")
	query.Set("state", state)
	if providerName == "google" {
		query.Set("scope", "openid email profile")
		query.Set("nonce", nonce)
	} else {
		query.Set("scope", "email,public_profile")
	}
	authorizationURL.RawQuery = query.Encode()
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "oauth login started",
		"content": map[string]string{"authorization_url": authorizationURL.String()},
	})
}

func (h *Handler) oauthCallback(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, enabled := h.service.provider(providerName)
	if !enabled {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	stateCookie, err := r.Cookie("alumnihub_oauth_state_" + providerName)
	state := r.URL.Query().Get("state")
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(state)) != 1 {
		h.redirectOAuthError(w, "oauth_state_invalid")
		return
	}
	nonce := ""
	if providerName == "google" {
		nonceCookie, nonceErr := r.Cookie("alumnihub_oauth_nonce_google")
		if nonceErr != nil || nonceCookie.Value == "" {
			h.redirectOAuthError(w, "oauth_state_invalid")
			return
		}
		nonce = nonceCookie.Value
	}
	http.SetCookie(w, &http.Cookie{
		Name: "alumnihub_oauth_state_" + providerName, Value: "",
		Path: "/api/v1/auth/oauth/" + providerName, MaxAge: -1,
		HttpOnly: true, Secure: h.service.cookieSecure, SameSite: http.SameSiteLaxMode,
	})
	if providerName == "google" {
		http.SetCookie(w, &http.Cookie{
			Name: "alumnihub_oauth_nonce_google", Value: "",
			Path: "/api/v1/auth/oauth/google", MaxAge: -1,
			HttpOnly: true, Secure: h.service.cookieSecure, SameSite: http.SameSiteLaxMode,
		})
	}
	if r.URL.Query().Get("error") != "" || r.URL.Query().Get("code") == "" {
		h.redirectOAuthError(w, "oauth_cancelled")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	user, err := exchangeOAuthCode(ctx, h.service.httpClient, h.service.googleVerifier, providerName, provider, r.URL.Query().Get("code"), nonce)
	if err != nil {
		h.reportFailure("oauth_provider_exchange", err)
		h.redirectOAuthError(w, "oauth_provider_failed")
		return
	}
	principal, token, expiresAt, err := h.service.loginOAuthUser(ctx, providerName, user)
	if err != nil {
		h.reportFailure("oauth_session", err)
		h.redirectOAuthError(w, "oauth_session_failed")
		return
	}
	h.service.SetSessionCookie(w, token, expiresAt)
	destination := "/"
	if principal.Role == RoleOnboarding {
		destination = "/registrasi/selesaikan/"
	}
	http.Redirect(w, r, strings.TrimRight(h.service.integrations.PublicBaseURL, "/")+destination, http.StatusSeeOther)
}

func (h *Handler) redirectOAuthError(w http.ResponseWriter, code string) {
	loginURL, err := url.Parse(strings.TrimRight(h.service.integrations.PublicBaseURL, "/") + "/login/")
	if err != nil || loginURL.Host == "" {
		writeError(w, http.StatusInternalServerError, "oauth login failed")
		return
	}
	query := loginURL.Query()
	query.Set("oauth_error", code)
	loginURL.RawQuery = query.Encode()
	w.Header().Set("Location", loginURL.String())
	w.WriteHeader(http.StatusSeeOther)
}

func (s *Service) verifyCaptcha(ctx context.Context, token, remoteIP string) error {
	if s.integrations.TurnstileSecret == "" || token == "" {
		return ErrCaptchaRejected
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	form := url.Values{
		"secret":   {s.integrations.TurnstileSecret},
		"response": {token},
		"remoteip": {remoteIP},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, turnstileVerifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create captcha request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := s.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCaptchaUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: provider returned status %d", ErrCaptchaUnavailable, response.StatusCode)
	}
	var result struct {
		Success  bool   `json:"success"`
		Hostname string `json:"hostname"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4096)).Decode(&result); err != nil {
		return fmt.Errorf("%w: invalid provider response", ErrCaptchaUnavailable)
	}
	if !result.Success {
		return ErrCaptchaRejected
	}
	publicURL, err := url.Parse(s.integrations.PublicBaseURL)
	if err != nil || result.Hostname != publicURL.Hostname() {
		return ErrCaptchaRejected
	}
	return nil
}

func exchangeOAuthCode(ctx context.Context, client *http.Client, googleVerifier googleIDTokenVerifier, providerName string, provider oauthProvider, code, nonce string) (providerUser, error) {
	form := url.Values{
		"client_id": {provider.clientID}, "client_secret": {provider.clientSecret},
		"code": {code}, "redirect_uri": {provider.redirectURI}, "grant_type": {"authorization_code"},
	}
	tokenRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return providerUser{}, fmt.Errorf("create oauth token request: %w", err)
	}
	tokenRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenResponse, err := client.Do(tokenRequest)
	if err != nil {
		return providerUser{}, fmt.Errorf("exchange oauth code: %w", err)
	}
	defer tokenResponse.Body.Close()
	if tokenResponse.StatusCode != http.StatusOK {
		return providerUser{}, errors.New("oauth token exchange rejected")
	}
	var token struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(io.LimitReader(tokenResponse.Body, 8192)).Decode(&token); err != nil || token.AccessToken == "" {
		return providerUser{}, errors.New("invalid oauth token response")
	}
	if providerName == "google" {
		if token.IDToken == "" || googleVerifier == nil {
			return providerUser{}, errors.New("google token response has no verifiable id token")
		}
		return googleVerifier.verify(ctx, client, provider.clientID, token.IDToken, nonce)
	}
	userInfoRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.userInfoURL, nil)
	if err != nil {
		return providerUser{}, fmt.Errorf("create oauth userinfo request: %w", err)
	}
	userInfoRequest.Header.Set("Authorization", "Bearer "+token.AccessToken)
	userInfoResponse, err := client.Do(userInfoRequest)
	if err != nil {
		return providerUser{}, fmt.Errorf("load oauth user profile: %w", err)
	}
	defer userInfoResponse.Body.Close()
	if userInfoResponse.StatusCode != http.StatusOK {
		return providerUser{}, errors.New("oauth userinfo request rejected")
	}
	var user struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(userInfoResponse.Body, 8192)).Decode(&user); err != nil {
		return providerUser{}, errors.New("invalid facebook userinfo response")
	}
	return providerUser{sub: user.ID, email: user.Email, name: user.Name}, nil
}

func (s *Service) loginOAuthUser(ctx context.Context, provider string, user providerUser) (Principal, string, time.Time, error) {
	if user.sub == "" || user.name == "" {
		return Principal{}, "", time.Time{}, errors.New("oauth identity is incomplete")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("begin oauth account transaction: %w", err)
	}
	defer tx.Rollback()
	var accountID []byte
	var name string
	err = tx.QueryRowContext(ctx, `
		SELECT a.id, a.name FROM external_identities e
		JOIN accounts a ON a.id = e.account_id
		WHERE e.provider = ? AND e.provider_subject = ? AND a.status = 'active'`,
		provider, user.sub,
	).Scan(&accountID, &name)
	if errors.Is(err, sql.ErrNoRows) {
		accountID, err = randomID()
		if err != nil {
			return Principal{}, "", time.Time{}, fmt.Errorf("generate oauth account ID: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO accounts (id, email, name, account_type, status) VALUES (?, NULL, ?, 0, 'active')`, accountID, user.name); err != nil {
			return Principal{}, "", time.Time{}, fmt.Errorf("create oauth account: %w", err)
		}
		identityID, err := randomID()
		if err != nil {
			return Principal{}, "", time.Time{}, fmt.Errorf("generate external identity ID: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO external_identities (id, account_id, provider, provider_subject) VALUES (?, ?, ?, ?)`, identityID, accountID, provider, user.sub); err != nil {
			return Principal{}, "", time.Time{}, fmt.Errorf("store external identity: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO account_roles (account_id, role_code) VALUES (?, 'member')`, accountID); err != nil {
			return Principal{}, "", time.Time{}, fmt.Errorf("assign member role: %w", err)
		}
		email := strings.ToLower(strings.TrimSpace(user.email))
		if parsed, parseErr := mail.ParseAddress(email); email != "" && parseErr == nil && parsed.Address == email && len(email) <= 254 {
			var contactExists bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM account_contacts WHERE value = ?)`, email).Scan(&contactExists); err != nil {
				return Principal{}, "", time.Time{}, fmt.Errorf("check oauth email ownership: %w", err)
			}
			if !contactExists {
				contactID, err := randomID()
				if err != nil {
					return Principal{}, "", time.Time{}, fmt.Errorf("generate contact ID: %w", err)
				}
				var verifiedAt any
				if user.emailVerified {
					verifiedAt = time.Now().UTC()
				}
				if _, err := tx.ExecContext(ctx, `
					INSERT INTO account_contacts (id, account_id, contact_type, value, is_primary, verified_at)
					VALUES (?, ?, 'email', ?, TRUE, ?)`,
					contactID, accountID, email, verifiedAt,
				); err != nil {
					return Principal{}, "", time.Time{}, fmt.Errorf("store oauth email contact: %w", err)
				}
			}
		}
		name = user.name
	} else if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("find oauth identity: %w", err)
	}
	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(rawToken)
	tokenHash := sha256.Sum256([]byte(token))
	sessionID, err := randomID()
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("generate session ID: %w", err)
	}
	expiresAt := s.now().UTC().Add(8 * time.Hour)
	if _, err := tx.ExecContext(ctx, `INSERT INTO sessions (id, account_id, token_hash, expires_at) VALUES (?, ?, ?, ?)`, sessionID, accountID, tokenHash[:], expiresAt); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("store oauth session: %w", err)
	}
	if err := insertAuditEvent(ctx, tx, accountID, sessionID, "auth.oauth_login"); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("audit oauth login: %w", err)
	}
	var roles, verificationStatus sql.NullString
	var email string
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(GROUP_CONCAT(DISTINCT ar.role_code ORDER BY ar.role_code), ''),
			COALESCE(ap.verification_status, ''), COALESCE(c.value, '')
		FROM accounts a
		LEFT JOIN account_roles ar ON ar.account_id = a.id
		LEFT JOIN alumni_profiles ap ON ap.account_id = a.id
		LEFT JOIN account_contacts c ON c.account_id = a.id AND c.contact_type = 'email' AND c.is_primary = TRUE
		WHERE a.id = ?
		GROUP BY a.id, ap.verification_status, c.value`, accountID,
	).Scan(&roles, &verificationStatus, &email); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("load oauth principal: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("commit oauth session: %w", err)
	}
	role := effectiveRole(roles.String, verificationStatus.String)
	return Principal{
		AccountID: formatID(accountID), Email: email, Name: name,
		Role: role, Capabilities: CapabilitiesFor(role), ProfileComplete: verificationStatus.Valid,
	}, token, expiresAt, nil
}

var (
	ErrCaptchaRejected    = errors.New("captcha was rejected")
	ErrCaptchaUnavailable = errors.New("captcha verification is unavailable")
)

func constantTimeEqual(left, right string) bool {
	return len(left) == len(right) && subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
