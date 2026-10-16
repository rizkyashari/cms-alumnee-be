package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
)

// requiredAttestations is the number of valid same-batch attestations a
// verification request needs before an Admin may approve it. Any active member
// of the same batch with an alumni profile (verified or not) is a classmate and
// may attest; the Admin's final decision is the safeguard against collusion.

const requiredAttestations = 5

const maxDecisionNoteRunes = 500

// validAttestationCount counts attestations that are still valid at decision
// time: the attester's account is active and they still belong to the same
// batch as the candidate.
const validAttestationCount = `
	SELECT COUNT(*) FROM alumni_attestations at
	JOIN accounts aa ON aa.id = at.attester_id AND aa.status = 'active'
	JOIN alumni_profiles atp ON atp.account_id = at.attester_id
	JOIN alumni_profiles tp ON tp.id = at.profile_id AND tp.batch_year = atp.batch_year
	WHERE at.profile_id = %s`

var (
	ErrVerificationNotFound     = errors.New("verification request not found")
	ErrVerificationPending      = errors.New("verification request is already pending")
	ErrAlreadyVerified          = errors.New("profile is already verified")
	ErrProfileRequired          = errors.New("alumni profile is required")
	ErrNotEligibleAttester      = errors.New("account is not an eligible attester for this alumnus")
	ErrSelfAttestation          = errors.New("an account cannot vouch for itself")
	ErrAlreadyAttested          = errors.New("attestation already recorded")
	ErrInsufficientAttestations = errors.New("not enough valid attestations")
	ErrBootstrapUnavailable     = errors.New("first-member confirmation is no longer available")
	ErrInvalidDecisionNote      = errors.New("decision note is invalid")
)

type verificationStatus struct {
	Status           string `json:"status"` // none | pending | approved | rejected
	Verified         bool   `json:"verified"`
	AttestationCount int    `json:"attestation_count"`
	Required         int    `json:"required_attestations"`
	DecisionNote     string `json:"decision_note,omitempty"`
}

type verificationCandidate struct {
	ProfileID        string         `json:"profile_id"`
	Name             string         `json:"name"`
	BatchYear        int            `json:"batch_year"`
	City             string         `json:"city"`
	RequestedAt      string         `json:"requested_at"`
	AttestationCount int            `json:"attestation_count"`
	Required         int            `json:"required_attestations"`
	Attesters        []attesterInfo `json:"attesters,omitempty"`
	AlreadyAttested  bool           `json:"already_attested"`
	CanAttest        bool           `json:"can_attest"`
	CanBootstrap     bool           `json:"can_bootstrap"` // first-member exception is still open
	IsSelf           bool           `json:"is_self"`
}

type attesterInfo struct {
	Name     string `json:"name"`
	Verified bool   `json:"verified"`
}

type callerProfile struct {
	id        []byte
	batchYear int
}

func (s *Service) loadCallerProfile(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, accountID []byte) (callerProfile, bool, error) {
	var profile callerProfile
	err := q.QueryRowContext(ctx, `
		SELECT id, batch_year FROM alumni_profiles WHERE account_id = ?`,
		accountID).Scan(&profile.id, &profile.batchYear)
	if errors.Is(err, sql.ErrNoRows) {
		return callerProfile{}, false, nil
	}
	if err != nil {
		return callerProfile{}, false, fmt.Errorf("load caller profile: %w", err)
	}
	return profile, true, nil
}

func (s *Service) ownVerification(ctx context.Context, accountID []byte) (verificationStatus, error) {
	result := verificationStatus{Status: "none", Required: requiredAttestations}
	var profileID []byte
	var profileStatus string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, verification_status FROM alumni_profiles WHERE account_id = ?`, accountID).Scan(&profileID, &profileStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return result, ErrProfileRequired
	}
	if err != nil {
		return result, fmt.Errorf("load verification profile: %w", err)
	}
	if profileStatus == "verified" {
		result.Status = "approved"
		result.Verified = true
		return result, nil
	}
	var note sql.NullString
	var requestStatus string
	err = s.db.QueryRowContext(ctx, `
		SELECT status, decision_note FROM verification_requests WHERE profile_id = ?`, profileID).Scan(&requestStatus, &note)
	if errors.Is(err, sql.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("load verification request: %w", err)
	}
	result.Status = requestStatus
	result.DecisionNote = note.String
	if err := s.db.QueryRowContext(ctx, fmt.Sprintf(validAttestationCount, "?"), profileID).Scan(&result.AttestationCount); err != nil {
		return result, fmt.Errorf("count attestations: %w", err)
	}
	return result, nil
}

// requestVerification opens (or reopens after a rejection) the member's
// verification request. The member cannot choose any status themselves.
func (s *Service) requestVerification(ctx context.Context, accountID []byte) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin verification request: %w", err)
	}
	defer tx.Rollback()
	var profileID []byte
	var profileStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT id, verification_status FROM alumni_profiles WHERE account_id = ? FOR UPDATE`, accountID).Scan(&profileID, &profileStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrProfileRequired
	}
	if err != nil {
		return fmt.Errorf("lock verification profile: %w", err)
	}
	if profileStatus == "verified" {
		return ErrAlreadyVerified
	}
	var existing string
	err = tx.QueryRowContext(ctx, `SELECT status FROM verification_requests WHERE profile_id = ? FOR UPDATE`, profileID).Scan(&existing)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = tx.ExecContext(ctx, `INSERT INTO verification_requests (profile_id) VALUES (?)`, profileID)
	case err != nil:
		return fmt.Errorf("load verification request: %w", err)
	case existing == "pending":
		return ErrVerificationPending
	default:
		_, err = tx.ExecContext(ctx, `
			UPDATE verification_requests
			SET status = 'pending', decision_kind = NULL, decided_by = NULL, decided_at = NULL,
				decision_note = NULL, requested_at = UTC_TIMESTAMP(6)
			WHERE profile_id = ?`, profileID)
	}
	if err != nil {
		return fmt.Errorf("store verification request: %w", err)
	}
	if err := insertAuditEventForResource(ctx, tx, accountID, "alumni_profile", profileID, "verification.request", []byte(`{}`)); err != nil {
		return fmt.Errorf("audit verification request: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit verification request: %w", err)
	}
	return nil
}

// listPending returns pending requests. Verified members only see their own
// batch; Pengurus and Admin see every pending request. The capability flags
// are advisory for the UI: the write endpoints re-check everything.
func (s *Service) listPending(ctx context.Context, principal Principal, withAttesters bool) ([]verificationCandidate, error) {
	accountID := principalIDBytes(principal.AccountID)
	caller, hasProfile, err := s.loadCallerProfile(ctx, s.db, accountID)
	if err != nil {
		return nil, err
	}
	staff := principal.Role == RoleOfficer || principal.Role == RoleAdmin
	bootstrapOpen := false
	if staff {
		var verified int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alumni_profiles WHERE verification_status = 'verified'`).Scan(&verified); err != nil {
			return nil, fmt.Errorf("count verified alumni: %w", err)
		}
		bootstrapOpen = verified == 0
	}
	query := `
		SELECT ap.id, ap.account_id, ap.full_name, ap.batch_year, ap.city, vr.requested_at,
			(` + fmt.Sprintf(validAttestationCount, "ap.id") + `),
			EXISTS(SELECT 1 FROM alumni_attestations mine WHERE mine.profile_id = ap.id AND mine.attester_id = ?),
			COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT('name', x.full_name, 'verified', IF(x.verification_status = 'verified', CAST('true' AS JSON), CAST('false' AS JSON))))
				FROM alumni_attestations xa
				JOIN alumni_profiles x ON x.account_id = xa.attester_id
				WHERE xa.profile_id = ap.id), JSON_ARRAY())
		FROM verification_requests vr
		JOIN alumni_profiles ap ON ap.id = vr.profile_id
		JOIN accounts a ON a.id = ap.account_id AND a.status = 'active'
		WHERE vr.status = 'pending'`
	args := []any{accountID}
	if !staff {
		if !hasProfile {
			return []verificationCandidate{}, nil
		}
		query += ` AND ap.batch_year = ?`
		args = append(args, caller.batchYear)
	}
	query += ` ORDER BY vr.requested_at LIMIT 100`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query pending verifications: %w", err)
	}
	defer rows.Close()
	items := make([]verificationCandidate, 0)
	for rows.Next() {
		var item verificationCandidate
		var profileID, ownerID []byte
		var requestedAt time.Time
		var attested bool
		var attesters string
		if err := rows.Scan(&profileID, &ownerID, &item.Name, &item.BatchYear, &item.City, &requestedAt,
			&item.AttestationCount, &attested, &attesters); err != nil {
			return nil, fmt.Errorf("scan pending verification: %w", err)
		}
		item.ProfileID = formatID(profileID)
		item.RequestedAt = requestedAt.UTC().Format(time.RFC3339)
		item.Required = requiredAttestations
		item.AlreadyAttested = attested
		item.IsSelf = string(ownerID) == string(accountID)
		if withAttesters {
			_ = json.Unmarshal([]byte(attesters), &item.Attesters)
		}
		item.CanAttest = hasProfile && caller.batchYear == item.BatchYear && !item.IsSelf && !attested
		item.CanBootstrap = bootstrapOpen && !item.IsSelf
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending verifications: %w", err)
	}
	return items, nil
}

// attest records one same-batch member's attestation for a pending
// request. The request row is locked, so concurrent attestations serialize.
func (s *Service) attest(ctx context.Context, attesterAccountID, profileID []byte) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin attestation: %w", err)
	}
	defer tx.Rollback()
	caller, hasProfile, err := s.loadCallerProfile(ctx, tx, attesterAccountID)
	if err != nil {
		return err
	}
	if !hasProfile {
		return ErrNotEligibleAttester
	}
	var targetAccount []byte
	var targetBatch int
	err = tx.QueryRowContext(ctx, `
		SELECT ap.account_id, ap.batch_year
		FROM alumni_profiles ap
		JOIN verification_requests vr ON vr.profile_id = ap.id AND vr.status = 'pending'
		JOIN accounts a ON a.id = ap.account_id AND a.status = 'active'
		WHERE ap.id = ? FOR UPDATE`, profileID).Scan(&targetAccount, &targetBatch)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrVerificationNotFound
	}
	if err != nil {
		return fmt.Errorf("lock verification request: %w", err)
	}
	if string(targetAccount) == string(attesterAccountID) {
		return ErrSelfAttestation
	}
	if targetBatch != caller.batchYear {
		return ErrNotEligibleAttester
	}
	var already bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM alumni_attestations WHERE profile_id = ? AND attester_id = ?)`,
		profileID, attesterAccountID).Scan(&already); err != nil {
		return fmt.Errorf("check existing attestation: %w", err)
	}
	if already {
		return ErrAlreadyAttested
	}
	attestationID, err := randomID()
	if err != nil {
		return fmt.Errorf("generate attestation ID: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO alumni_attestations (id, profile_id, attester_id) VALUES (?, ?, ?)`,
		attestationID, profileID, attesterAccountID); err != nil {
		return fmt.Errorf("store attestation: %w", err)
	}
	if err := insertAuditEventForResource(ctx, tx, attesterAccountID, "alumni_profile", profileID, "verification.attest", []byte(`{}`)); err != nil {
		return fmt.Errorf("audit attestation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit attestation: %w", err)
	}
	return nil
}

// decide lets an Admin approve (only with enough valid attestations) or reject
// a pending request. Approval flips the single source of truth,
// alumni_profiles.verification_status.
func (s *Service) decide(ctx context.Context, actorID, profileID []byte, approve bool, note string) error {
	note = strings.TrimSpace(note)
	if utf8.RuneCountInString(note) > maxDecisionNoteRunes {
		return ErrInvalidDecisionNote
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin verification decision: %w", err)
	}
	defer tx.Rollback()
	var targetAccount []byte
	err = tx.QueryRowContext(ctx, `
		SELECT ap.account_id FROM verification_requests vr
		JOIN alumni_profiles ap ON ap.id = vr.profile_id
		WHERE vr.profile_id = ? AND vr.status = 'pending' FOR UPDATE`, profileID).Scan(&targetAccount)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrVerificationNotFound
	}
	if err != nil {
		return fmt.Errorf("lock verification request: %w", err)
	}
	action, status := "verification.reject", "rejected"
	if approve {
		if string(targetAccount) == string(actorID) {
			return ErrSelfAttestation
		}
		var count int
		if err := tx.QueryRowContext(ctx, fmt.Sprintf(validAttestationCount, "?"), profileID).Scan(&count); err != nil {
			return fmt.Errorf("count attestations: %w", err)
		}
		if count < requiredAttestations {
			return ErrInsufficientAttestations
		}
		action, status = "verification.approve", "approved"
	}
	if err := s.finishDecision(ctx, tx, actorID, profileID, status, "peer", note, action, approve); err != nil {
		return err
	}
	return tx.Commit()
}

// bootstrapConfirm is the one-time first-member exception: while no verified
// alumnus exists, a Pengurus or Admin may confirm one pending member without
// attestations. A row lock on the 'admin' role serializes concurrent calls so
// the exception cannot be used twice.
func (s *Service) bootstrapConfirm(ctx context.Context, actorID, profileID []byte, note string) error {
	note = strings.TrimSpace(note)
	if utf8.RuneCountInString(note) > maxDecisionNoteRunes {
		return ErrInvalidDecisionNote
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin first-member confirmation: %w", err)
	}
	defer tx.Rollback()
	var lock string
	if err := tx.QueryRowContext(ctx, `SELECT code FROM roles WHERE code = 'admin' FOR UPDATE`).Scan(&lock); err != nil {
		return fmt.Errorf("lock first-member confirmation: %w", err)
	}
	var verified int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM alumni_profiles WHERE verification_status = 'verified'`).Scan(&verified); err != nil {
		return fmt.Errorf("count verified alumni: %w", err)
	}
	if verified > 0 {
		return ErrBootstrapUnavailable
	}
	var targetAccount []byte
	err = tx.QueryRowContext(ctx, `
		SELECT ap.account_id FROM verification_requests vr
		JOIN alumni_profiles ap ON ap.id = vr.profile_id
		WHERE vr.profile_id = ? AND vr.status = 'pending' FOR UPDATE`, profileID).Scan(&targetAccount)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrVerificationNotFound
	}
	if err != nil {
		return fmt.Errorf("lock verification request: %w", err)
	}
	if string(targetAccount) == string(actorID) {
		return ErrSelfAttestation
	}
	if err := s.finishDecision(ctx, tx, actorID, profileID, "approved", "bootstrap", note, "verification.bootstrap_approve", true); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) finishDecision(ctx context.Context, tx *sql.Tx, actorID, profileID []byte, status, kind, note, action string, verify bool) error {
	var noteValue any
	if note != "" {
		noteValue = note
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE verification_requests
		SET status = ?, decision_kind = ?, decided_by = ?, decided_at = UTC_TIMESTAMP(6), decision_note = ?
		WHERE profile_id = ?`, status, kind, actorID, noteValue, profileID); err != nil {
		return fmt.Errorf("store verification decision: %w", err)
	}
	if verify {
		if _, err := tx.ExecContext(ctx, `
			UPDATE alumni_profiles SET verification_status = 'verified' WHERE id = ?`, profileID); err != nil {
			return fmt.Errorf("verify alumni profile: %w", err)
		}
	}
	metadata, err := json.Marshal(map[string]string{"decision": status, "kind": kind})
	if err != nil {
		return fmt.Errorf("encode verification audit metadata: %w", err)
	}
	if err := insertAuditEventForResource(ctx, tx, actorID, "alumni_profile", profileID, action, metadata); err != nil {
		return fmt.Errorf("audit verification decision: %w", err)
	}
	return nil
}

type verificationAuditEntry struct {
	Action    string `json:"action"`
	Kind      string `json:"kind,omitempty"`
	Actor     string `json:"actor"`
	Subject   string `json:"subject"`
	CreatedAt string `json:"created_at"`
}

// listVerificationAudit returns the latest verification audit events so an
// Admin can review who attested, approved, rejected or bootstrapped whom.
func (s *Service) listVerificationAudit(ctx context.Context) ([]verificationAuditEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ae.action, ae.created_at, COALESCE(actor.name, ''), COALESCE(ap.full_name, ''),
			COALESCE(JSON_UNQUOTE(JSON_EXTRACT(ae.metadata, '$.kind')), '')
		FROM audit_events ae
		LEFT JOIN accounts actor ON actor.id = ae.actor_id
		LEFT JOIN alumni_profiles ap ON ap.id = ae.resource_id
		WHERE ae.action LIKE 'verification.%'
		ORDER BY ae.created_at DESC, ae.id DESC
		LIMIT 100`)
	if err != nil {
		return nil, fmt.Errorf("query verification audit: %w", err)
	}
	defer rows.Close()
	entries := make([]verificationAuditEntry, 0)
	for rows.Next() {
		var entry verificationAuditEntry
		var createdAt time.Time
		if err := rows.Scan(&entry.Action, &createdAt, &entry.Actor, &entry.Subject, &entry.Kind); err != nil {
			return nil, fmt.Errorf("scan verification audit: %w", err)
		}
		entry.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate verification audit: %w", err)
	}
	return entries, nil
}

func (h *Handler) verificationAudit(w http.ResponseWriter, r *http.Request) {
	entries, err := h.service.listVerificationAudit(r.Context())
	if err != nil {
		h.reportFailure("verification_audit", err)
		writeError(w, http.StatusInternalServerError, "could not load audit trail")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{"message": "ok", "content": map[string]any{"items": entries}})
}

func (h *Handler) registerVerificationRoutes(router chi.Router) {
	router.With(h.Authenticate).Get("/api/v1/verification/me", h.myVerification)
	router.With(h.Authenticate).Post("/api/v1/verification/request", h.requestVerification)
	router.With(h.RequireCapability(CapabilityAttest)).Get("/api/v1/verification/candidates", h.verificationCandidates)
	router.With(h.RequireCapability(CapabilityAttest)).Post("/api/v1/verification/{profileID}/attest", h.attestProfile)
	router.With(h.RequireCapability(CapabilityAttest)).Post("/api/v1/verification/{profileID}/bootstrap", h.bootstrapProfile)
	router.With(h.RequireCapability(CapabilityVerifyAlumni)).Get("/api/v1/admin/verifications", h.adminVerifications)
	router.With(h.RequireCapability(CapabilityVerifyAlumni)).Get("/api/v1/admin/verifications/audit", h.verificationAudit)
	router.With(h.RequireCapability(CapabilityVerifyAlumni)).Post("/api/v1/admin/verifications/{profileID}/{decision}", h.decideVerification)
	router.With(h.Authenticate).Get("/api/v1/auth/profile/photo", h.ownPhoto)
	router.With(h.RequireCapability(CapabilityAttest)).Get("/api/v1/verification/{profileID}/photo", h.reviewPhoto)
}

func (h *Handler) myVerification(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	status, err := h.service.ownVerification(r.Context(), principalIDBytes(principal.AccountID))
	if errors.Is(err, ErrProfileRequired) {
		writeError(w, http.StatusConflict, "complete the alumni profile first")
		return
	}
	if err != nil {
		h.reportFailure("own_verification", err)
		writeError(w, http.StatusInternalServerError, "could not load verification status")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{"message": "ok", "content": status})
}

func (h *Handler) requestVerification(w http.ResponseWriter, r *http.Request) {
	if !h.service.sameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	principal, _ := PrincipalFromContext(r.Context())
	err := h.service.requestVerification(r.Context(), principalIDBytes(principal.AccountID))
	switch {
	case errors.Is(err, ErrProfileRequired):
		writeError(w, http.StatusConflict, "complete the alumni profile first")
	case errors.Is(err, ErrAlreadyVerified):
		writeError(w, http.StatusConflict, "profile is already verified")
	case errors.Is(err, ErrVerificationPending):
		writeError(w, http.StatusConflict, "verification request is already pending")
	case err != nil:
		h.reportFailure("request_verification", err)
		writeError(w, http.StatusInternalServerError, "could not submit verification request")
	default:
		writeJSON(w, http.StatusAccepted, map[string]any{"message": "verification requested", "content": nil})
	}
}

func (h *Handler) verificationCandidates(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	items, err := h.service.listPending(r.Context(), principal, false)
	if err != nil {
		h.reportFailure("list_verification_candidates", err)
		writeError(w, http.StatusInternalServerError, "could not load verification requests")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{"message": "ok", "content": map[string]any{"items": items}})
}

func (h *Handler) adminVerifications(w http.ResponseWriter, r *http.Request) {
	principal, _ := PrincipalFromContext(r.Context())
	items, err := h.service.listPending(r.Context(), principal, true)
	if err != nil {
		h.reportFailure("list_admin_verifications", err)
		writeError(w, http.StatusInternalServerError, "could not load verification requests")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{"message": "ok", "content": map[string]any{"items": items}})
}

func (h *Handler) attestProfile(w http.ResponseWriter, r *http.Request) {
	if !h.service.sameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	profileID := principalIDBytes(chi.URLParam(r, "profileID"))
	principal, _ := PrincipalFromContext(r.Context())
	if profileID == nil {
		writeError(w, http.StatusNotFound, "verification request not found")
		return
	}
	err := h.service.attest(r.Context(), principalIDBytes(principal.AccountID), profileID)
	switch {
	case errors.Is(err, ErrVerificationNotFound):
		writeError(w, http.StatusNotFound, "verification request not found")
	case errors.Is(err, ErrSelfAttestation):
		writeError(w, http.StatusForbidden, "you cannot vouch for yourself")
	case errors.Is(err, ErrNotEligibleAttester):
		writeError(w, http.StatusForbidden, "only members of the same batch can vouch for this alumnus")
	case errors.Is(err, ErrAlreadyAttested):
		writeError(w, http.StatusConflict, "you have already vouched for this alumnus")
	case err != nil:
		h.reportFailure("attest", err)
		writeError(w, http.StatusInternalServerError, "could not record attestation")
	default:
		writeJSON(w, http.StatusCreated, map[string]any{"message": "attestation recorded", "content": nil})
	}
}

func (h *Handler) bootstrapProfile(w http.ResponseWriter, r *http.Request) {
	if !h.service.sameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	principal, _ := PrincipalFromContext(r.Context())
	if principal.Role != RoleOfficer && principal.Role != RoleAdmin {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	profileID := principalIDBytes(chi.URLParam(r, "profileID"))
	if profileID == nil {
		writeError(w, http.StatusNotFound, "verification request not found")
		return
	}
	var input struct {
		Note string `json:"note"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	err := h.service.bootstrapConfirm(r.Context(), principalIDBytes(principal.AccountID), profileID, input.Note)
	h.writeDecisionResult(w, err, "first alumnus confirmed", "bootstrap_verification")
}

func (h *Handler) decideVerification(w http.ResponseWriter, r *http.Request) {
	if !h.service.sameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	decision := chi.URLParam(r, "decision")
	if decision != "approve" && decision != "reject" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	profileID := principalIDBytes(chi.URLParam(r, "profileID"))
	if profileID == nil {
		writeError(w, http.StatusNotFound, "verification request not found")
		return
	}
	var input struct {
		Note string `json:"note"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	principal, _ := PrincipalFromContext(r.Context())
	err := h.service.decide(r.Context(), principalIDBytes(principal.AccountID), profileID, decision == "approve", input.Note)
	message := map[string]string{"approve": "verification approved", "reject": "verification rejected"}[decision]
	h.writeDecisionResult(w, err, message, "decide_verification")
}

func (h *Handler) writeDecisionResult(w http.ResponseWriter, err error, message, operation string) {
	switch {
	case errors.Is(err, ErrVerificationNotFound):
		writeError(w, http.StatusNotFound, "no pending verification request was found")
	case errors.Is(err, ErrSelfAttestation):
		writeError(w, http.StatusForbidden, "you cannot decide your own verification")
	case errors.Is(err, ErrInsufficientAttestations):
		writeError(w, http.StatusConflict, fmt.Sprintf("at least %d valid attestations from same-batch members are required", requiredAttestations))
	case errors.Is(err, ErrBootstrapUnavailable):
		writeError(w, http.StatusConflict, "first-member confirmation is only available while no alumnus is verified")
	case errors.Is(err, ErrInvalidDecisionNote):
		writeError(w, http.StatusBadRequest, "decision note is too long")
	case err != nil:
		h.reportFailure(operation, err)
		writeError(w, http.StatusInternalServerError, "could not record decision")
	default:
		writeJSON(w, http.StatusOK, map[string]any{"message": message, "content": nil})
	}
}
