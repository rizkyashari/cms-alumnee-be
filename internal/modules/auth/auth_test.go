package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestCapabilitiesForEffectiveRoles(t *testing.T) {
	tests := []struct {
		role Role
		want []Capability
	}{
		{RolePublic, []Capability{CapabilityPublicContent}},
		{RoleNonVerified, []Capability{CapabilityPublicContent, CapabilityDirectoryLite, CapabilityProfileEdit, CapabilityBusinessSubmit}},
		{RoleVerified, []Capability{CapabilityPublicContent, CapabilityDirectoryFull, CapabilityProfileEdit, CapabilityBusinessSubmit, CapabilityComment, CapabilityAttest}},
		{RoleOfficer, []Capability{CapabilityPublicContent, CapabilityDocumentManage, CapabilityAttest}},
		{RoleAdmin, []Capability{CapabilityPublicContent, CapabilityDirectoryFull, CapabilityProfileEdit, CapabilityBusinessSubmit, CapabilityComment, CapabilityDocumentManage, CapabilityAttest, CapabilityAdminManage, CapabilityVerifyAlumni}},
	}
	for _, test := range tests {
		t.Run(string(test.role), func(t *testing.T) {
			got := CapabilitiesFor(test.role)
			if len(got) != len(test.want) {
				t.Fatalf("CapabilitiesFor(%q) returned %v, want %v", test.role, got, test.want)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Fatalf("CapabilitiesFor(%q) returned %v, want %v", test.role, got, test.want)
				}
			}
		})
	}
}

func TestRoleCapabilityMatrix(t *testing.T) {
	matrix := map[Capability]map[Role]bool{
		CapabilityDirectoryLite: {
			RolePublic: false, RoleNonVerified: true, RoleVerified: false, RoleOfficer: false, RoleAdmin: false,
		},
		CapabilityProfileEdit: {
			RolePublic: false, RoleNonVerified: true, RoleVerified: true, RoleOfficer: false, RoleAdmin: true,
		},
		CapabilityBusinessSubmit: {
			RolePublic: false, RoleNonVerified: true, RoleVerified: true, RoleOfficer: false, RoleAdmin: true,
		},
		CapabilityDirectoryFull: {
			RolePublic: false, RoleNonVerified: false, RoleVerified: true, RoleOfficer: false, RoleAdmin: true,
		},
		CapabilityComment: {
			RolePublic: false, RoleNonVerified: false, RoleVerified: true, RoleOfficer: false, RoleAdmin: true,
		},
		CapabilityDocumentManage: {
			RolePublic: false, RoleNonVerified: false, RoleVerified: false, RoleOfficer: true, RoleAdmin: true,
		},
		CapabilityAttest: {
			RolePublic: false, RoleNonVerified: false, RoleVerified: true, RoleOfficer: true, RoleAdmin: true,
		},
		CapabilityAdminManage: {
			RolePublic: false, RoleNonVerified: false, RoleVerified: false, RoleOfficer: false, RoleAdmin: true,
		},
		CapabilityVerifyAlumni: {
			RolePublic: false, RoleNonVerified: false, RoleVerified: false, RoleOfficer: false, RoleAdmin: true,
		},
	}
	roles := []Role{RolePublic, RoleNonVerified, RoleVerified, RoleOfficer, RoleAdmin}
	for capability, roleRules := range matrix {
		for _, role := range roles {
			granted := hasCapability(CapabilitiesFor(role), capability)
			if granted != roleRules[role] {
				t.Errorf("role %q capability %q = %t, want %t", role, capability, granted, roleRules[role])
			}
		}
	}
}

func TestEffectiveRoleUsesPrivilegedRoleAndVerificationStatus(t *testing.T) {
	tests := []struct {
		roles  string
		status string
		want   Role
	}{
		{"", "", RolePublic},
		{"member", "non_verified", RoleNonVerified},
		{"member", "verified", RoleVerified},
		{"member,officer", "verified", RoleOfficer},
		{"admin,member,officer", "non_verified", RoleAdmin},
	}
	for _, test := range tests {
		if got := effectiveRole(test.roles, test.status); got != test.want {
			t.Errorf("effectiveRole(%q, %q) = %q, want %q", test.roles, test.status, got, test.want)
		}
	}
}

func TestRequireCapabilityRejectsAnonymousAndUnauthorizedRoles(t *testing.T) {
	protected := RequireCapability(CapabilityComment)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	anonymous := httptest.NewRecorder()
	protected.ServeHTTP(anonymous, httptest.NewRequest(http.MethodGet, "/", nil))
	if anonymous.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d, want %d", anonymous.Code, http.StatusUnauthorized)
	}

	nonVerified := httptest.NewRequest(http.MethodGet, "/", nil)
	nonVerified = nonVerified.WithContext(withPrincipal(nonVerified, Principal{
		Role: RoleNonVerified, Capabilities: CapabilitiesFor(RoleNonVerified),
	}))
	forbidden := httptest.NewRecorder()
	protected.ServeHTTP(forbidden, nonVerified)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("non-verified status = %d, want %d", forbidden.Code, http.StatusForbidden)
	}

	verified := httptest.NewRequest(http.MethodGet, "/", nil)
	verified = verified.WithContext(withPrincipal(verified, Principal{
		Role: RoleVerified, Capabilities: CapabilitiesFor(RoleVerified),
	}))
	allowed := httptest.NewRecorder()
	protected.ServeHTTP(allowed, verified)
	if allowed.Code != http.StatusNoContent {
		t.Fatalf("verified status = %d, want %d", allowed.Code, http.StatusNoContent)
	}
}

func TestCapabilityMiddlewareEnforcesEveryEffectiveRole(t *testing.T) {
	roles := []Role{RolePublic, RoleNonVerified, RoleVerified, RoleOfficer, RoleAdmin}
	capabilities := []Capability{
		CapabilityPublicContent,
		CapabilityDirectoryLite,
		CapabilityDirectoryFull,
		CapabilityProfileEdit,
		CapabilityBusinessSubmit,
		CapabilityComment,
		CapabilityDocumentManage,
		CapabilityAttest,
		CapabilityAdminManage,
		CapabilityVerifyAlumni,
		Capability("unknown"),
	}
	spoofedCapabilities := append([]Capability(nil), capabilities...)

	for _, role := range roles {
		for _, capability := range capabilities {
			t.Run(string(role)+"/"+string(capability), func(t *testing.T) {
				next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})
				request := httptest.NewRequest(http.MethodGet, "/", nil)
				expectedStatus := http.StatusForbidden
				request = request.WithContext(withPrincipal(request, Principal{
					Role: role, Capabilities: spoofedCapabilities,
				}))
				if roleHasCapability(role, capability) {
					expectedStatus = http.StatusNoContent
				}

				recorder := httptest.NewRecorder()
				RequireCapability(capability)(next).ServeHTTP(recorder, request)
				if recorder.Code != expectedStatus {
					t.Fatalf("role %q capability %q returned %d, want %d", role, capability, recorder.Code, expectedStatus)
				}
			})
		}
	}
}

func TestCapabilityMiddlewareRejectsSpoofedAndUnknownPrincipals(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	tests := []Principal{
		{Role: RoleNonVerified, Capabilities: []Capability{CapabilityAdminManage, CapabilityVerifyAlumni}},
		{Role: Role("unknown"), Capabilities: []Capability{CapabilityAdminManage, CapabilityVerifyAlumni}},
		{Role: RolePublic, Capabilities: []Capability{CapabilityAdminManage, CapabilityVerifyAlumni}},
	}
	for _, principal := range tests {
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request = request.WithContext(withPrincipal(request, principal))
		recorder := httptest.NewRecorder()
		RequireCapability(CapabilityVerifyAlumni)(next).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusForbidden {
			t.Errorf("principal role %q with supplied capabilities returned %d, want %d", principal.Role, recorder.Code, http.StatusForbidden)
		}
	}
}

func TestAnyCapabilityMiddlewareEnforcesPublicAndAuthenticatedAccess(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	tests := []struct {
		name         string
		role         Role
		capabilities []Capability
		want         int
	}{
		{name: "anonymous public content", capabilities: []Capability{CapabilityPublicContent}, want: http.StatusNoContent},
		{name: "anonymous private content", capabilities: []Capability{CapabilityDirectoryFull}, want: http.StatusUnauthorized},
		{name: "non-verified limited directory", role: RoleNonVerified, capabilities: []Capability{CapabilityDirectoryLite, CapabilityDirectoryFull}, want: http.StatusNoContent},
		{name: "non-verified cannot access full directory", role: RoleNonVerified, capabilities: []Capability{CapabilityDirectoryFull}, want: http.StatusForbidden},
		{name: "officer authorized documents", role: RoleOfficer, capabilities: []Capability{CapabilityDocumentManage, CapabilityVerifyAlumni}, want: http.StatusNoContent},
		{name: "officer cannot verify", role: RoleOfficer, capabilities: []Capability{CapabilityVerifyAlumni}, want: http.StatusForbidden},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.role != "" {
				request = request.WithContext(withPrincipal(request, Principal{
					Role: test.role, Capabilities: []Capability{CapabilityPublicContent},
				}))
			}
			recorder := httptest.NewRecorder()
			RequireAnyCapability(test.capabilities...)(next).ServeHTTP(recorder, request)
			if recorder.Code != test.want {
				t.Fatalf("status = %d, want %d", recorder.Code, test.want)
			}
		})
	}
}

func TestHandlerCapabilityMiddlewareAllowsPublicAndRejectsAnonymousPrivateRequests(t *testing.T) {
	handler := NewHandler(NewService(nil, false, ""), nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	tests := []struct {
		name       string
		middleware func(http.Handler) http.Handler
		want       int
	}{
		{name: "public content", middleware: handler.RequireCapability(CapabilityPublicContent), want: http.StatusNoContent},
		{name: "private profile", middleware: handler.RequireCapability(CapabilityProfileEdit), want: http.StatusUnauthorized},
		{name: "public or directory", middleware: handler.RequireAnyCapability(CapabilityPublicContent, CapabilityDirectoryFull), want: http.StatusNoContent},
		{name: "directory only", middleware: handler.RequireAnyCapability(CapabilityDirectoryFull), want: http.StatusUnauthorized},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			test.middleware(next).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
			if recorder.Code != test.want {
				t.Fatalf("status = %d, want %d", recorder.Code, test.want)
			}
		})
	}
}

func TestAuthenticationAndCapabilityMiddlewareRejectAnonymousRequests(t *testing.T) {
	handler := NewHandler(NewService(nil, false, ""), nil)
	router := chi.NewRouter()
	router.With(handler.RequireCapability(CapabilityProfileEdit)).Get("/profile", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/profile", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func withPrincipal(r *http.Request, principal Principal) context.Context {
	return context.WithValue(r.Context(), principalContextKey{}, principal)
}

func hasCapability(capabilities []Capability, want Capability) bool {
	for _, capability := range capabilities {
		if capability == want {
			return true
		}
	}
	return false
}

func TestSessionCookieIsHttpOnlyAndSameSite(t *testing.T) {
	service := NewService(nil, true, "test-local-password")
	recorder := httptest.NewRecorder()
	service.SetSessionCookie(recorder, "test-token", time.Now().Add(time.Hour))
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if !cookie.HttpOnly || cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/api/v1" {
		t.Fatalf("unexpected local cookie flags: %+v", cookie)
	}
}

func TestDemoLoginAndCaptchaAreDisabledWithoutExplicitConfiguration(t *testing.T) {
	service := NewService(nil, false, "")
	if _, _, _, err := service.DemoLogin(context.Background(), LoginInput{
		Email: "admin@demo.local", Password: "local-test-password",
	}); err != ErrDemoAuthDisabled {
		t.Fatalf("DemoLogin() error = %v, want %v", err, ErrDemoAuthDisabled)
	}
	if err := service.verifyCaptcha(context.Background(), "", "127.0.0.1"); err != ErrCaptchaRejected {
		t.Fatalf("verifyCaptcha() error = %v, want %v", err, ErrCaptchaRejected)
	}
}

func TestLogoutRejectsCrossOriginRequest(t *testing.T) {
	handler := NewHandler(NewService(nil, false, ""), nil)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	request := httptest.NewRequest(http.MethodPost, "http://alumni.example/api/v1/auth/logout", nil)
	request.Header.Set("Origin", "https://attacker.example")
	request.Header.Set("X-Forwarded-Host", "attacker.example")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("cross-origin logout status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestSameOriginValidationAllowsProxiedLoopbackOriginForDemo(t *testing.T) {
	service := NewService(nil, true, "local-demo-password")
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:3000/api/v1/auth/logout", nil)
	request.Header.Set("Origin", "http://localhost:3001")
	if !service.sameOriginRequest(request) {
		t.Fatal("expected loopback frontend origin to be allowed through local API proxy")
	}
}

func TestSameOriginValidationUsesConfiguredProductionOrigin(t *testing.T) {
	service := NewService(nil, false, "")
	service.ConfigureIntegrations(Integrations{PublicBaseURL: "https://alumni.example"})
	request := httptest.NewRequest(http.MethodPost, "http://api:3000/api/v1/auth/logout", nil)
	request.Header.Set("Origin", "https://alumni.example")
	if !service.sameOriginRequest(request) {
		t.Fatal("expected configured frontend origin to be allowed through reverse proxy")
	}
	request.Header.Set("Origin", "https://attacker.example")
	if service.sameOriginRequest(request) {
		t.Fatal("unconfigured external origin should be rejected")
	}
}

func TestDemoLoginRejectsNonLoopbackHostAndOrigin(t *testing.T) {
	handler := NewHandler(NewService(nil, true, "local-demo-password"), nil)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	tests := []struct {
		name   string
		target string
		origin string
	}{
		{name: "public host", target: "https://alumni.example/api/v1/auth/demo-login"},
		{name: "public origin", target: "http://localhost/api/v1/auth/demo-login", origin: "https://alumni.example"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.target, nil)
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusNotFound {
				t.Fatalf("demo login status = %d, want %d", recorder.Code, http.StatusNotFound)
			}
		})
	}
}

func TestLoginLimiterEnforcesWindowLimit(t *testing.T) {
	limiter := newLoginLimiter(2, time.Minute)
	limiter.now = func() time.Time { return time.Unix(100, 0) }
	if !limiter.Allow("127.0.0.1") || !limiter.Allow("127.0.0.1") {
		t.Fatal("first two login attempts should be allowed")
	}
	if limiter.Allow("127.0.0.1") {
		t.Fatal("login attempt beyond configured limit should be rejected")
	}
}
