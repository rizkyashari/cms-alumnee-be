package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

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
		{RoleNonVerified, false, false},
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
	service := &Service{}
	domestic := func(code string) registrationProfile {
		return registrationProfile{countryCode: "ID", cityCode: code}
	}
	if !service.regencyCodeKnown(domestic("99.99")) {
		t.Fatal("format-only validation should accept any well-formed code when no master list is loaded")
	}
	service.SetRegencyCodes(codes)
	if !service.regencyCodeKnown(domestic("34.04")) {
		t.Fatal("known regency rejected")
	}
	if service.regencyCodeKnown(domestic("99.99")) {
		t.Fatal("unknown regency accepted")
	}
	if !service.regencyCodeKnown(registrationProfile{countryCode: "US", city: "Seattle"}) {
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
