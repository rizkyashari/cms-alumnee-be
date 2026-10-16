package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestVerificationRoutesRequireAuthentication(t *testing.T) {
	router := chi.NewRouter()
	NewHandler(&Service{}, nil).RegisterRoutes(router)
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/verification/me"},
		{http.MethodPost, "/api/v1/verification/request"},
		{http.MethodGet, "/api/v1/verification/candidates"},
		{http.MethodPost, "/api/v1/verification/00000000-0000-4000-8000-000000000000/attest"},
		{http.MethodPost, "/api/v1/verification/00000000-0000-4000-8000-000000000000/bootstrap"},
		{http.MethodGet, "/api/v1/verification/00000000-0000-4000-8000-000000000000/photo"},
		{http.MethodGet, "/api/v1/admin/verifications"},
		{http.MethodPost, "/api/v1/admin/verifications/00000000-0000-4000-8000-000000000000/approve"},
		{http.MethodGet, "/api/v1/auth/profile/photo"},
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(route.method, route.path, nil))
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", route.method, route.path, recorder.Code)
		}
	}
}

func TestVerificationCapabilitiesMatchPolicy(t *testing.T) {
	for _, test := range []struct {
		role       Role
		attest     bool
		verifyFlow bool
	}{
		{RoleNonVerified, true, false},
		{RoleVerified, true, false},
		{RoleOfficer, true, false},
		{RoleAdmin, true, true},
	} {
		if got := roleHasCapability(test.role, CapabilityAttest); got != test.attest {
			t.Errorf("%s attest = %v, want %v", test.role, got, test.attest)
		}
		if got := roleHasCapability(test.role, CapabilityVerifyAlumni); got != test.verifyFlow {
			t.Errorf("%s verify = %v, want %v", test.role, got, test.verifyFlow)
		}
	}
}

func TestRegencyMasterValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "regencies.json")
	if err := os.WriteFile(path, []byte(`[{"id":"34.04","name":"KABUPATEN SLEMAN"},{"id":"34.71","name":"KOTA YOGYAKARTA"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	codes, err := LoadRegencyCodes(path)
	if err != nil {
		t.Fatalf("LoadRegencyCodes: %v", err)
	}
	unavailable := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	defer unavailable.Close()
	service := &Service{regionBaseURL: unavailable.URL, now: time.Now}
	ctx := context.Background()
	domestic := func(code string) registrationProfile {
		return registrationProfile{countryCode: "ID", cityCode: code}
	}
	if !service.regencyCodeKnown(ctx, domestic("99.99")) {
		t.Fatal("format-only validation should accept any well-formed code when no master list is loaded")
	}
	service.SetRegencyCodes(codes)
	if !service.regencyCodeKnown(ctx, domestic("34.04")) {
		t.Fatal("known regency rejected")
	}
	if service.regencyCodeKnown(ctx, domestic("99.99")) {
		t.Fatal("unknown regency accepted")
	}
	if !service.regencyCodeKnown(ctx, registrationProfile{countryCode: "US", city: "Seattle"}) {
		t.Fatal("international location should not be checked against Indonesian master data")
	}
	for name, content := range map[string]string{"empty": `[]`, "bad code": `[{"id":"3404"}]`, "not json": `x`} {
		bad := filepath.Join(t.TempDir(), name+".json")
		if err := os.WriteFile(bad, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadRegencyCodes(bad); err == nil {
			t.Errorf("LoadRegencyCodes accepted %s", name)
		}
	}
}

func TestRegionProxyServesValidatesAndCaches(t *testing.T) {
	var hits int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		switch r.URL.Path {
		case "/provinces.json":
			_, _ = w.Write([]byte(`{"data":[{"id":"34","name":"DI YOGYAKARTA"},{"id":"bad","name":"X"},{"id":"31","name":"DKI JAKARTA"}]}`))
		case "/regencies/34.json":
			_, _ = w.Write([]byte(`{"data":[{"id":"34.04","name":"KABUPATEN SLEMAN"},{"id":"34.71","name":"KOTA YOGYAKARTA"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer upstream.Close()
	now := time.Now()
	service := &Service{regionBaseURL: upstream.URL, httpClient: upstream.Client(), now: func() time.Time { return now }}
	router := chi.NewRouter()
	NewHandler(service, nil).RegisterRoutes(router)

	get := func(path string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		return recorder
	}
	response := get("/api/v1/regions/provinces")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "DKI JAKARTA") || strings.Contains(response.Body.String(), "bad") {
		t.Fatalf("provinces = %d %s", response.Code, response.Body.String())
	}
	if response = get("/api/v1/regions/provinces/34/regencies"); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "34.04") {
		t.Fatalf("regencies = %d %s", response.Code, response.Body.String())
	}
	before := hits
	get("/api/v1/regions/provinces")
	get("/api/v1/regions/provinces/34/regencies")
	if hits != before {
		t.Fatalf("cache not used: %d upstream hits after warm requests", hits-before)
	}
	if code := get("/api/v1/regions/provinces/abc/regencies").Code; code != http.StatusNotFound {
		t.Fatalf("invalid province code = %d, want 404", code)
	}
	ctx := context.Background()
	if !service.regencyCodeKnown(ctx, registrationProfile{countryCode: "ID", cityCode: "34.04"}) {
		t.Fatal("known regency rejected")
	}
	if service.regencyCodeKnown(ctx, registrationProfile{countryCode: "ID", cityCode: "34.99"}) {
		t.Fatal("unknown regency accepted")
	}
	// A stale entry is still served if the provider goes away.
	now = now.Add(2 * regionCacheTTL)
	upstream.Close()
	if response = get("/api/v1/regions/provinces"); response.Code != http.StatusOK {
		t.Fatalf("stale fallback = %d, want 200", response.Code)
	}
}

func TestRegionProxyReportsUnavailableWithoutCache(t *testing.T) {
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadGateway) }))
	defer down.Close()
	router := chi.NewRouter()
	NewHandler(&Service{regionBaseURL: down.URL, httpClient: down.Client(), now: time.Now}, nil).RegisterRoutes(router)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/regions/provinces", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", recorder.Code)
	}
}
