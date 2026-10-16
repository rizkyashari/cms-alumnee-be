package auth

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rizkyashari/cms-alumnee-be/internal/platform/migrations"
)

// These tests need a disposable MySQL 8 database. Set TEST_MYSQL_DSN, for
// example "user:pass@tcp(127.0.0.1:3399)/dbname?parseTime=true". They delete
// all accounts, so never point them at a shared database.
func openVerificationDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := migrations.Up(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	for _, statement := range []string{`DELETE FROM audit_events`, `DELETE FROM accounts`} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("reset database: %v", err)
		}
	}
	return db
}

type seededMember struct {
	accountID []byte
	profileID []byte
}

func seedMember(t *testing.T, db *sql.DB, name string, batch int, status string, roles ...string) seededMember {
	t.Helper()
	accountID, _ := randomID()
	profileID, _ := randomID()
	if _, err := db.Exec(`INSERT INTO accounts (id, name) VALUES (?, ?)`, accountID, name); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if len(roles) == 0 {
		roles = []string{"member"}
	}
	for _, role := range roles {
		if _, err := db.Exec(`INSERT INTO account_roles (account_id, role_code) VALUES (?, ?)`, accountID, role); err != nil {
			t.Fatalf("seed role: %v", err)
		}
	}
	if _, err := db.Exec(`
		INSERT INTO alumni_profiles (id, account_id, full_name, batch_year, city, verification_status)
		VALUES (?, ?, ?, ?, 'Yogyakarta', ?)`, profileID, accountID, name, batch, status); err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	return seededMember{accountID: accountID, profileID: profileID}
}

func verificationStatusOf(t *testing.T, s *Service, m seededMember) verificationStatus {
	t.Helper()
	status, err := s.ownVerification(context.Background(), m.accountID)
	if err != nil {
		t.Fatalf("ownVerification: %v", err)
	}
	return status
}

func TestPeerVerificationWorkflow(t *testing.T) {
	db := openVerificationDB(t)
	s := &Service{db: db}
	ctx := context.Background()
	admin := seedMember(t, db, "Admin", 2000, "non_verified", "admin")
	candidate := seedMember(t, db, "Calon", 2012, "non_verified")
	// Classmates need not be verified themselves: three verified, two not.
	peers := make([]seededMember, 5)
	for i := range peers {
		status := "verified"
		if i >= 3 {
			status = "non_verified"
		}
		peers[i] = seedMember(t, db, "Teman", 2012, status)
	}
	otherBatch := seedMember(t, db, "Angkatan Lain", 2011, "verified")
	noProfile := seedMember(t, db, "Admin Tanpa Profil", 2012, "non_verified", "officer")
	if _, err := db.Exec(`DELETE FROM alumni_profiles WHERE account_id = ?`, noProfile.accountID); err != nil {
		t.Fatal(err)
	}

	if got := verificationStatusOf(t, s, candidate); got.Status != "none" {
		t.Fatalf("initial status = %q, want none", got.Status)
	}
	if err := s.attest(ctx, peers[0].accountID, candidate.profileID); !errors.Is(err, ErrVerificationNotFound) {
		t.Fatalf("attest without request = %v, want ErrVerificationNotFound", err)
	}
	if err := s.requestVerification(ctx, candidate.accountID); err != nil {
		t.Fatalf("requestVerification: %v", err)
	}
	if err := s.requestVerification(ctx, candidate.accountID); !errors.Is(err, ErrVerificationPending) {
		t.Fatalf("duplicate request = %v, want ErrVerificationPending", err)
	}
	if err := s.requestVerification(ctx, peers[0].accountID); !errors.Is(err, ErrAlreadyVerified) {
		t.Fatalf("verified request = %v, want ErrAlreadyVerified", err)
	}

	if err := s.attest(ctx, otherBatch.accountID, candidate.profileID); !errors.Is(err, ErrNotEligibleAttester) {
		t.Fatalf("other batch attest = %v, want ErrNotEligibleAttester", err)
	}
	if err := s.attest(ctx, noProfile.accountID, candidate.profileID); !errors.Is(err, ErrNotEligibleAttester) {
		t.Fatalf("attest without alumni profile = %v, want ErrNotEligibleAttester", err)
	}
	if err := s.attest(ctx, candidate.accountID, candidate.profileID); !errors.Is(err, ErrSelfAttestation) {
		t.Fatalf("self attest = %v, want ErrSelfAttestation", err)
	}
	for i := 0; i < 4; i++ {
		if err := s.attest(ctx, peers[i].accountID, candidate.profileID); err != nil {
			t.Fatalf("attest %d: %v", i, err)
		}
	}
	if err := s.attest(ctx, peers[0].accountID, candidate.profileID); !errors.Is(err, ErrAlreadyAttested) {
		t.Fatalf("duplicate attest = %v, want ErrAlreadyAttested", err)
	}
	if err := s.decide(ctx, admin.accountID, candidate.profileID, true, ""); !errors.Is(err, ErrInsufficientAttestations) {
		t.Fatalf("approve with 4 attestations = %v, want ErrInsufficientAttestations", err)
	}
	if err := s.attest(ctx, peers[4].accountID, candidate.profileID); err != nil {
		t.Fatalf("fifth attest: %v", err)
	}
	if got := verificationStatusOf(t, s, candidate); got.AttestationCount != 5 || got.Status != "pending" {
		t.Fatalf("status = %+v, want pending with 5 attestations", got)
	}
	// An attester whose account is disabled no longer counts.
	if _, err := db.Exec(`UPDATE accounts SET status = 'disabled' WHERE id = ?`, peers[4].accountID); err != nil {
		t.Fatal(err)
	}
	if err := s.decide(ctx, admin.accountID, candidate.profileID, true, ""); !errors.Is(err, ErrInsufficientAttestations) {
		t.Fatalf("approve after attester disabled = %v, want ErrInsufficientAttestations", err)
	}
	if _, err := db.Exec(`UPDATE accounts SET status = 'active' WHERE id = ?`, peers[4].accountID); err != nil {
		t.Fatal(err)
	}

	pending, err := s.listPending(ctx, Principal{AccountID: formatID(peers[0].accountID), Role: RoleVerified}, false)
	if err != nil || len(pending) != 1 || !pending[0].AlreadyAttested || pending[0].CanAttest {
		t.Fatalf("verified peer queue = %+v, %v", pending, err)
	}
	if pending, err = s.listPending(ctx, Principal{AccountID: formatID(otherBatch.accountID), Role: RoleVerified}, false); err != nil || len(pending) != 0 {
		t.Fatalf("other batch queue = %+v, %v; want empty", pending, err)
	}
	if pending, err = s.listPending(ctx, Principal{AccountID: formatID(admin.accountID), Role: RoleAdmin}, true); err != nil || len(pending) != 1 || len(pending[0].Attesters) != 5 || pending[0].Attesters[0].Name == "" {
		t.Fatalf("admin queue = %+v, %v; want one request with five attesters", pending, err)
	}

	if err := s.decide(ctx, admin.accountID, candidate.profileID, true, "ok"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if got := verificationStatusOf(t, s, candidate); !got.Verified || got.Status != "approved" {
		t.Fatalf("status after approval = %+v", got)
	}
	if err := s.decide(ctx, admin.accountID, candidate.profileID, true, ""); !errors.Is(err, ErrVerificationNotFound) {
		t.Fatalf("second decision = %v, want ErrVerificationNotFound", err)
	}
	var audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_events WHERE action IN ('verification.request','verification.attest','verification.approve')`).Scan(&audits); err != nil || audits != 7 {
		t.Fatalf("audit events = %d, %v; want 7", audits, err)
	}
}

func TestRejectedRequestCanBeResubmitted(t *testing.T) {
	db := openVerificationDB(t)
	s := &Service{db: db}
	ctx := context.Background()
	admin := seedMember(t, db, "Admin", 2000, "non_verified", "admin")
	candidate := seedMember(t, db, "Calon", 2013, "non_verified")
	if err := s.requestVerification(ctx, candidate.accountID); err != nil {
		t.Fatal(err)
	}
	if err := s.decide(ctx, admin.accountID, candidate.profileID, false, "data tidak cocok"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	got := verificationStatusOf(t, s, candidate)
	if got.Status != "rejected" || got.DecisionNote != "data tidak cocok" {
		t.Fatalf("status = %+v, want rejected with note", got)
	}
	if err := s.requestVerification(ctx, candidate.accountID); err != nil {
		t.Fatalf("resubmit: %v", err)
	}
	if got := verificationStatusOf(t, s, candidate); got.Status != "pending" || got.DecisionNote != "" {
		t.Fatalf("status after resubmit = %+v", got)
	}
}

func TestBootstrapConfirmIsOneTimeAndRaceSafe(t *testing.T) {
	db := openVerificationDB(t)
	s := &Service{db: db}
	ctx := context.Background()
	officer := seedMember(t, db, "Pengurus", 2000, "non_verified", "officer", "member")
	first := seedMember(t, db, "Pertama", 2012, "non_verified")
	second := seedMember(t, db, "Kedua", 2012, "non_verified")
	for _, m := range []seededMember{first, second} {
		if err := s.requestVerification(ctx, m.accountID); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.bootstrapConfirm(ctx, first.accountID, first.profileID, ""); !errors.Is(err, ErrSelfAttestation) {
		t.Fatalf("self bootstrap = %v, want ErrSelfAttestation", err)
	}
	var wg sync.WaitGroup
	results := make([]error, 2)
	for i, m := range []seededMember{first, second} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = s.bootstrapConfirm(ctx, officer.accountID, m.profileID, "konfirmasi awal")
		}()
	}
	wg.Wait()
	successes := 0
	for _, err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrBootstrapUnavailable) {
			t.Fatalf("unexpected bootstrap error: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent bootstrap successes = %d, want exactly 1 (%v)", successes, results)
	}
	var verified int
	if err := db.QueryRow(`SELECT COUNT(*) FROM alumni_profiles WHERE verification_status = 'verified'`).Scan(&verified); err != nil || verified != 1 {
		t.Fatalf("verified profiles = %d, %v; want 1", verified, err)
	}
	var kind string
	if err := db.QueryRow(`SELECT decision_kind FROM verification_requests WHERE status = 'approved'`).Scan(&kind); err != nil || kind != "bootstrap" {
		t.Fatalf("decision kind = %q, %v; want bootstrap", kind, err)
	}
}

func testPhoto(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestPhotoEndpointsRestrictAccess(t *testing.T) {
	db := openVerificationDB(t)
	s := &Service{db: db}
	h := &Handler{service: s}
	ctx := context.Background()
	candidate := seedMember(t, db, "Calon", 2012, "non_verified")
	peer := seedMember(t, db, "Teman", 2012, "verified")
	otherBatch := seedMember(t, db, "Lain", 2011, "verified")
	officer := seedMember(t, db, "Pengurus", 2000, "non_verified", "officer")
	photo := testPhoto(t)
	if _, err := db.Exec(`INSERT INTO alumni_profile_photos (account_id, mime_type, image_data) VALUES (?, 'image/png', ?)`, candidate.accountID, photo); err != nil {
		t.Fatal(err)
	}

	router := chi.NewRouter()
	var principal Principal
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalContextKey{}, principal)))
		})
	})
	router.Get("/photo/me", h.ownPhoto)
	router.Get("/photo/{profileID}", h.reviewPhoto)
	get := func(path string, as Principal) *httptest.ResponseRecorder {
		principal = as
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		return recorder
	}

	owner := Principal{AccountID: formatID(candidate.accountID), Role: RoleNonVerified}
	response := get("/photo/me", owner)
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "image/png" ||
		response.Header().Get("X-Content-Type-Options") != "nosniff" || !bytes.Equal(response.Body.Bytes(), photo) {
		t.Fatalf("own photo: status %d, headers %v", response.Code, response.Header())
	}
	if code := get("/photo/me", Principal{AccountID: formatID(peer.accountID), Role: RoleVerified}).Code; code != http.StatusNotFound {
		t.Fatalf("member without photo = %d, want 404", code)
	}

	path := "/photo/" + formatID(candidate.profileID)
	if code := get(path, Principal{AccountID: formatID(peer.accountID), Role: RoleVerified}).Code; code != http.StatusNotFound {
		t.Fatalf("photo without pending request = %d, want 404", code)
	}
	if err := s.requestVerification(ctx, candidate.accountID); err != nil {
		t.Fatal(err)
	}
	if code := get(path, Principal{AccountID: formatID(peer.accountID), Role: RoleVerified}).Code; code != http.StatusOK {
		t.Fatalf("same-batch verified peer = %d, want 200", code)
	}
	if code := get(path, Principal{AccountID: formatID(officer.accountID), Role: RoleOfficer}).Code; code != http.StatusOK {
		t.Fatalf("officer = %d, want 200", code)
	}
	if code := get(path, Principal{AccountID: formatID(otherBatch.accountID), Role: RoleVerified}).Code; code != http.StatusNotFound {
		t.Fatalf("other-batch verified member = %d, want 404", code)
	}
	if code := get("/photo/not-an-id", Principal{AccountID: formatID(officer.accountID), Role: RoleOfficer}).Code; code != http.StatusNotFound {
		t.Fatalf("malformed id = %d, want 404", code)
	}
}

var testScenarioUsernames = []string{
	"admin.uji", "pengurus.satu", "pengurus.dua", "alumni.satu", "alumni.dua",
	"alumni.tiga", "alumni.empat", "alumni.lima", "alumni.enam", "alumni.tujuh",
}

func TestScenarioAccountsAndAuditTrail(t *testing.T) {
	db := openVerificationDB(t)
	s := &Service{db: db, now: time.Now}
	ctx := context.Background()
	if err := s.ProvisionScenarioAccounts(ctx, "short", testScenarioUsernames, false); err == nil {
		t.Fatal("short scenario password accepted")
	}
	if err := s.ProvisionScenarioAccounts(ctx, "", testScenarioUsernames, false); err != nil {
		t.Fatalf("empty password should be a no-op: %v", err)
	}
	if err := s.ProvisionScenarioAccounts(ctx, "scenario-password-1", testScenarioUsernames, false); err != nil {
		t.Fatalf("provision: %v", err)
	}
	if err := s.ProvisionScenarioAccounts(ctx, "scenario-password-1", testScenarioUsernames, false); err != nil {
		t.Fatalf("provision is not idempotent: %v", err)
	}
	var credentials, members int
	if err := db.QueryRow(`SELECT COUNT(*) FROM password_credentials WHERE is_enabled`).Scan(&credentials); err != nil || credentials != len(scenarioAccounts) {
		t.Fatalf("enabled credentials = %d, %v", credentials, err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM alumni_profiles WHERE batch_year = ? AND verification_status = 'non_verified'`, scenarioBatchYear).Scan(&members); err != nil || members != 7 {
		t.Fatalf("scenario members = %d, %v; want 7", members, err)
	}
	principal := func(username string) (Principal, []byte) {
		var id []byte
		if err := db.QueryRow(`SELECT account_id FROM password_credentials WHERE username = ?`, username).Scan(&id); err != nil {
			t.Fatal(err)
		}
		p, err := s.loadAccountPrincipal(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		return p, id
	}
	if p, _ := principal("admin.uji"); p.Role != RoleAdmin {
		t.Fatalf("admin.uji role = %s", p.Role)
	}
	if p, _ := principal("pengurus.dua"); p.Role != RoleOfficer {
		t.Fatalf("pengurus.dua role = %s", p.Role)
	}

	// Full walkthrough: first member via bootstrap, then a candidate with five attesters.
	_, officer := principal("pengurus.satu")
	_, admin := principal("admin.uji")
	_, first := principal("alumni.satu")
	_, second := principal("alumni.dua")
	profileOf := func(account []byte) []byte {
		var id []byte
		if err := db.QueryRow(`SELECT id FROM alumni_profiles WHERE account_id = ?`, account).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	for _, account := range [][]byte{first, second} {
		if err := s.requestVerification(ctx, account); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.bootstrapConfirm(ctx, officer, profileOf(first), "alumni pertama"); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	for _, username := range []string{"alumni.satu", "alumni.tiga", "alumni.empat", "alumni.lima", "alumni.enam"} {
		_, attester := principal(username)
		if err := s.attest(ctx, attester, profileOf(second)); err != nil {
			t.Fatalf("%s attest: %v", username, err)
		}
	}
	if err := s.decide(ctx, admin, profileOf(second), true, ""); err != nil {
		t.Fatalf("approve: %v", err)
	}
	entries, err := s.listVerificationAudit(ctx)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	actions := map[string]int{}
	for _, entry := range entries {
		actions[entry.Action]++
		if entry.Actor == "" || entry.Subject == "" {
			t.Fatalf("audit entry missing actor or subject: %+v", entry)
		}
	}
	if actions["verification.request"] != 2 || actions["verification.bootstrap_approve"] != 1 ||
		actions["verification.attest"] != 5 || actions["verification.approve"] != 1 {
		t.Fatalf("audit actions = %v", actions)
	}

	// Reset returns the cast to a clean state.
	if err := s.ProvisionScenarioAccounts(ctx, "scenario-password-1", testScenarioUsernames, true); err != nil {
		t.Fatalf("reset: %v", err)
	}
	var verified, attestations int
	_ = db.QueryRow(`SELECT COUNT(*) FROM alumni_profiles WHERE verification_status = 'verified'`).Scan(&verified)
	_ = db.QueryRow(`SELECT COUNT(*) FROM alumni_attestations`).Scan(&attestations)
	if verified != 0 || attestations != 0 {
		t.Fatalf("after reset: %d verified, %d attestations", verified, attestations)
	}
}
