package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const SessionCookieName = "alumnihub_session"

type Role string

const (
	RolePublic      Role = "public"
	RoleNonVerified Role = "non_verified"
	RoleVerified    Role = "verified"
	RoleOfficer     Role = "officer"
	RoleAdmin       Role = "admin"
)

type Capability string

const (
	CapabilityPublicContent  Capability = "content:read_public"
	CapabilityDirectoryLite  Capability = "directory:read_limited"
	CapabilityDirectoryFull  Capability = "directory:read_full"
	CapabilityProfileEdit    Capability = "profile:edit_self"
	CapabilityBusinessSubmit Capability = "business:submit"
	CapabilityComment        Capability = "content:comment"
	CapabilityDocumentManage Capability = "documents:manage"
	CapabilityAttest         Capability = "alumni:attest"
	CapabilityAdminManage    Capability = "admin:manage"
	CapabilityVerifyAlumni   Capability = "alumni:verify"
)

type Principal struct {
	AccountID    string       `json:"account_id"`
	Email        string       `json:"email"`
	Name         string       `json:"name"`
	Role         Role         `json:"role"`
	Capabilities []Capability `json:"capabilities"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Service struct {
	db              *sql.DB
	httpClient      *http.Client
	googleVerifier  googleIDTokenVerifier
	demoAuthEnabled bool
	demoPassword    [32]byte
	cookieSecure    bool
	integrations    Integrations
	now             func() time.Time
}

func NewService(db *sql.DB, demoAuthEnabled bool, demoPassword string) *Service {
	passwordHash := sha256.Sum256([]byte(demoPassword))
	return &Service{
		db:              db,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
		googleVerifier:  googleOIDCVerifier{},
		demoAuthEnabled: demoAuthEnabled,
		demoPassword:    passwordHash,
		cookieSecure:    !demoAuthEnabled,
		now:             time.Now,
	}
}

type Integrations struct {
	PublicBaseURL    string
	TurnstileSiteKey string
	TurnstileSecret  string
	GoogleClientID   string
	GoogleSecret     string
	FacebookClientID string
	FacebookSecret   string
}

func (s *Service) ConfigureIntegrations(integrations Integrations) {
	s.integrations = integrations
	if integrations.PublicBaseURL != "" {
		s.cookieSecure = strings.HasPrefix(integrations.PublicBaseURL, "https://")
	}
}

func CapabilitiesFor(role Role) []Capability {
	capabilities := []Capability{CapabilityPublicContent}
	switch role {
	case RoleNonVerified:
		return append(capabilities, CapabilityDirectoryLite, CapabilityProfileEdit, CapabilityBusinessSubmit)
	case RoleVerified:
		return append(capabilities, CapabilityDirectoryFull, CapabilityProfileEdit, CapabilityBusinessSubmit, CapabilityComment, CapabilityAttest)
	case RoleOfficer:
		return append(capabilities, CapabilityDocumentManage, CapabilityAttest)
	case RoleAdmin:
		return append(capabilities, CapabilityDirectoryFull, CapabilityProfileEdit, CapabilityBusinessSubmit, CapabilityComment, CapabilityDocumentManage, CapabilityAttest, CapabilityAdminManage, CapabilityVerifyAlumni)
	default:
		return capabilities
	}
}

func (s *Service) DemoLogin(ctx context.Context, input LoginInput) (Principal, string, time.Time, error) {
	if !s.demoAuthEnabled {
		return Principal{}, "", time.Time{}, ErrDemoAuthDisabled
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	demo, ok := demoAccounts[email]
	passwordHash := sha256.Sum256([]byte(input.Password))
	passwordMatches := subtle.ConstantTimeCompare(passwordHash[:], s.demoPassword[:]) == 1
	if !ok || !passwordMatches || input.Password == "" {
		return Principal{}, "", time.Time{}, ErrInvalidCredentials
	}

	principal, err := s.ensureDemoAccount(ctx, email, demo)
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("prepare demo account: %w", err)
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("begin session transaction: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO sessions (id, account_id, token_hash, expires_at) VALUES (?, ?, ?, ?)`,
		sessionID, principalIDBytes(principal.AccountID), tokenHash[:], expiresAt,
	); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("store session: %w", err)
	}
	if err := insertAuditEvent(ctx, tx, principalIDBytes(principal.AccountID), sessionID, "auth.demo_login"); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("audit demo login: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("commit session: %w", err)
	}
	return principal, token, expiresAt, nil
}

func (s *Service) Authenticate(r *http.Request) (Principal, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return Principal{}, ErrUnauthenticated
	}
	tokenHash := sha256.Sum256([]byte(cookie.Value))
	var accountID []byte
	var name, email, accountStatus string
	var roles, verificationStatus sql.NullString
	err = s.db.QueryRowContext(r.Context(), `
		SELECT a.id, a.name, COALESCE(c.value, ''), a.status,
			GROUP_CONCAT(DISTINCT ar.role_code ORDER BY ar.role_code),
			ap.verification_status
		FROM sessions s
		JOIN accounts a ON a.id = s.account_id
		LEFT JOIN account_contacts c ON c.account_id = a.id AND c.contact_type = 'email' AND c.is_primary = TRUE
		LEFT JOIN account_roles ar ON ar.account_id = a.id
		LEFT JOIN alumni_profiles ap ON ap.account_id = a.id
		WHERE s.token_hash = ? AND s.revoked_at IS NULL AND s.expires_at > UTC_TIMESTAMP(6)
		GROUP BY a.id, a.name, c.value, a.status, ap.verification_status`,
		tokenHash[:],
	).Scan(&accountID, &name, &email, &accountStatus, &roles, &verificationStatus)
	if errors.Is(err, sql.ErrNoRows) || accountStatus != "active" {
		return Principal{}, ErrUnauthenticated
	}
	if err != nil {
		return Principal{}, fmt.Errorf("load session principal: %w", err)
	}
	role := effectiveRole(roles.String, verificationStatus.String)
	if role == RolePublic {
		return Principal{}, ErrUnauthenticated
	}
	return Principal{
		AccountID:    formatID(accountID),
		Email:        email,
		Name:         name,
		Role:         role,
		Capabilities: CapabilitiesFor(role),
	}, nil
}

func (s *Service) RevokeSession(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	tokenHash := sha256.Sum256([]byte(token))
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin logout transaction: %w", err)
	}
	defer tx.Rollback()
	var accountID, sessionID []byte
	err = tx.QueryRowContext(ctx, `SELECT account_id, id FROM sessions WHERE token_hash = ? AND revoked_at IS NULL`, tokenHash[:]).Scan(&accountID, &sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load session for logout: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET revoked_at = UTC_TIMESTAMP(6) WHERE id = ? AND revoked_at IS NULL`, sessionID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if err := insertAuditEvent(ctx, tx, accountID, sessionID, "auth.logout"); err != nil {
		return fmt.Errorf("audit logout: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit logout: %w", err)
	}
	return nil
}

func (s *Service) SetSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/api/v1",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/api/v1",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

type demoAccount struct {
	name         string
	role         Role
	verification string
	batchYear    int
}

var demoAccounts = map[string]demoAccount{
	"nonverified@demo.local": {name: "Alumni Belum Terverifikasi", role: RoleNonVerified, verification: "non_verified", batchYear: 2012},
	"verified@demo.local":    {name: "Alumni Terverifikasi", role: RoleVerified, verification: "verified", batchYear: 2010},
	"officer@demo.local":     {name: "Pengurus Demo", role: RoleOfficer},
	"admin@demo.local":       {name: "Admin Demo", role: RoleAdmin},
}

func (s *Service) ensureDemoAccount(ctx context.Context, email string, demo demoAccount) (Principal, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Principal{}, fmt.Errorf("begin account transaction: %w", err)
	}
	defer tx.Rollback()

	var accountID []byte
	var accountStatus, assignedRoles, verificationStatus string
	var hasVerifiedContact bool
	err = tx.QueryRowContext(ctx, `
		SELECT a.id, a.status,
			COALESCE(GROUP_CONCAT(DISTINCT ar.role_code ORDER BY ar.role_code), ''),
			COALESCE(ap.verification_status, ''),
			MAX(c.verified_at IS NOT NULL)
		FROM accounts a
		JOIN account_contacts c ON c.account_id = a.id
		LEFT JOIN account_roles ar ON ar.account_id = a.id
		LEFT JOIN alumni_profiles ap ON ap.account_id = a.id
		WHERE c.contact_type = 'email' AND c.value = ?
		GROUP BY a.id, a.status, ap.verification_status`, email,
	).Scan(&accountID, &accountStatus, &assignedRoles, &verificationStatus, &hasVerifiedContact)
	roleCode := demoRoleCode(demo)
	if errors.Is(err, sql.ErrNoRows) {
		accountID, err = randomID()
		if err != nil {
			return Principal{}, fmt.Errorf("generate account ID: %w", err)
		}
		contactID, err := randomID()
		if err != nil {
			return Principal{}, fmt.Errorf("generate contact ID: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO accounts (id, email, name, account_type, status) VALUES (?, ?, ?, 0, 'active')`, accountID, email, demo.name); err != nil {
			return Principal{}, fmt.Errorf("create demo account: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO account_contacts (id, account_id, contact_type, value, is_primary, verified_at) VALUES (?, ?, 'email', ?, TRUE, UTC_TIMESTAMP(6))`, contactID, accountID, email); err != nil {
			return Principal{}, fmt.Errorf("create demo contact: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO account_roles (account_id, role_code) VALUES (?, ?)`, accountID, roleCode); err != nil {
			return Principal{}, fmt.Errorf("assign demo role: %w", err)
		}
		if demo.verification != "" {
			profileID, err := randomID()
			if err != nil {
				return Principal{}, fmt.Errorf("generate demo profile ID: %w", err)
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO alumni_profiles (id, account_id, full_name, batch_year, city, verification_status)
				VALUES (?, ?, ?, ?, 'Yogyakarta', ?)`, profileID, accountID, demo.name, demo.batchYear, demo.verification); err != nil {
				return Principal{}, fmt.Errorf("create demo profile: %w", err)
			}
		}
	} else if err != nil {
		return Principal{}, fmt.Errorf("find demo account: %w", err)
	} else if accountStatus != "active" || assignedRoles != roleCode ||
		verificationStatus != demo.verification || !hasVerifiedContact {
		return Principal{}, errors.New("existing demo account does not match its expected role and status")
	}
	if err := tx.Commit(); err != nil {
		return Principal{}, fmt.Errorf("commit demo account: %w", err)
	}

	return Principal{
		AccountID:    formatID(accountID),
		Email:        email,
		Name:         demo.name,
		Role:         demo.role,
		Capabilities: CapabilitiesFor(demo.role),
	}, nil
}

func demoRoleCode(demo demoAccount) string {
	switch demo.role {
	case RoleOfficer:
		return "officer"
	case RoleAdmin:
		return "admin"
	default:
		return "member"
	}
}

func effectiveRole(roles, verificationStatus string) Role {
	roleSet := make(map[string]bool)
	for _, role := range strings.Split(roles, ",") {
		roleSet[role] = true
	}
	if roleSet["admin"] {
		return RoleAdmin
	}
	if roleSet["officer"] {
		return RoleOfficer
	}
	if roleSet["member"] {
		if verificationStatus == "verified" {
			return RoleVerified
		}
		return RoleNonVerified
	}
	return RolePublic
}

func randomID() ([]byte, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return nil, err
	}
	id[6] = id[6]&0x0f | 0x40
	id[8] = id[8]&0x3f | 0x80
	return id, nil
}

func formatID(id []byte) string {
	if len(id) != 16 {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[:4], id[4:6], id[6:8], id[8:10], id[10:])
}

func principalIDBytes(id string) []byte {
	compact := strings.ReplaceAll(id, "-", "")
	result, err := hex.DecodeString(compact)
	if err != nil || len(result) != 16 {
		return nil
	}
	return result
}

func insertAuditEvent(ctx context.Context, tx *sql.Tx, actorID, resourceID []byte, action string) error {
	auditID, err := randomID()
	if err != nil {
		return fmt.Errorf("generate audit ID: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO audit_events (id, actor_id, action, resource_type, resource_id, metadata)
		VALUES (?, ?, ?, 'session', ?, JSON_OBJECT())`, auditID, actorID, action, resourceID)
	return err
}

var (
	ErrDemoAuthDisabled   = errors.New("demo authentication is disabled")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthenticated    = errors.New("authentication required")
	ErrForbidden          = errors.New("forbidden")
)
