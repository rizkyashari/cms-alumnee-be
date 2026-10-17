package auth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math/big"
	"mime"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	emailOTPValidity          = 10 * time.Minute
	registrationTokenValidity = 20 * time.Minute
	maxProfilePhotoBytes      = 2 << 20
)

var (
	domesticCityCodePattern = regexp.MustCompile(`^[0-9]{2}\.[0-9]{2}$`)
	countryCodePattern      = regexp.MustCompile(`^[A-Z]{2}$`)
)

type registrationProfile struct {
	fullName    string
	batchYear   int
	countryCode string
	cityCode    string
	city        string
	photo       []byte
	mimeType    string
}

type emailOTPVerification struct {
	principal  *Principal
	token      string
	tokenEmail string
	session    string
	expiresAt  time.Time
}

func (s *Service) emailOTPConfigured() bool {
	return s.integrations.ResendAPIKey != "" && s.integrations.ResendFromEmail != ""
}

func (s *Service) startEmailOTP(ctx context.Context, rawEmail string) error {
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 {
		return ErrInvalidRegistration
	}
	if !s.emailOTPConfigured() {
		return ErrEmailDeliveryUnavailable
	}

	var purpose string
	var accountID []byte
	var contactVerified sql.NullTime
	err = s.db.QueryRowContext(ctx, `
		SELECT account_id, verified_at FROM account_contacts
		WHERE contact_type = 'email' AND value = ?`, email).Scan(&accountID, &contactVerified)
	if err == nil {
		purpose = "login"
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check registration email: %w", err)
	} else if email == s.integrations.BootstrapAdminEmail {
		var adminCount int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM account_roles WHERE role_code = 'admin'`).Scan(&adminCount); err != nil {
			return fmt.Errorf("check initial administrator: %w", err)
		}
		if adminCount == 0 {
			purpose = "admin_bootstrap"
		}
	}
	if purpose == "" {
		purpose = "registration"
	}

	codeNumber, err := rand.Int(rand.Reader, bigOTPRange)
	if err != nil {
		return fmt.Errorf("generate email OTP: %w", err)
	}
	code := fmt.Sprintf("%06d", codeNumber.Int64())
	challengeID, err := randomID()
	if err != nil {
		return fmt.Errorf("generate OTP challenge ID: %w", err)
	}
	codeHash := s.emailOTPHash(challengeID, code)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin email OTP challenge: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		UPDATE email_otp_challenges SET consumed_at = UTC_TIMESTAMP(6)
		WHERE email = ? AND consumed_at IS NULL`, email); err != nil {
		return fmt.Errorf("expire previous email OTP challenges: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO email_otp_challenges (id, email, code_hash, purpose, expires_at)
		VALUES (?, ?, ?, ?, ?)`, challengeID, email, codeHash, purpose, s.now().UTC().Add(emailOTPValidity)); err != nil {
		return fmt.Errorf("store email OTP challenge: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit email OTP challenge: %w", err)
	}
	if err := s.sendEmailOTP(ctx, email, code); err != nil {
		_, revokeErr := s.db.ExecContext(ctx, `
			UPDATE email_otp_challenges SET consumed_at = UTC_TIMESTAMP(6)
			WHERE id = ? AND consumed_at IS NULL`, challengeID)
		if revokeErr != nil {
			return fmt.Errorf("email OTP delivery failed and challenge cleanup failed: %w", revokeErr)
		}
		return err
	}
	return nil
}

var bigOTPRange = mustOTPRange()

func mustOTPRange() *big.Int {
	return big.NewInt(1000000)
}

func (s *Service) emailOTPHash(challengeID []byte, code string) []byte {
	mac := hmac.New(sha256.New, []byte(s.integrations.ResendAPIKey))
	_, _ = mac.Write(challengeID)
	_, _ = mac.Write([]byte(code))
	return mac.Sum(nil)
}

func (s *Service) sendEmailOTP(ctx context.Context, email, code string) error {
	body := struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Text    string   `json:"text"`
	}{
		From:    s.integrations.ResendFromEmail,
		To:      []string{email},
		Subject: "Kode verifikasi AlumniHub",
		Text:    fmt.Sprintf("Kode OTP Anda: %s\nKode berlaku selama 10 menit. Jangan bagikan kode ini kepada siapa pun.", code),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode email OTP: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create email OTP request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+s.integrations.ResendAPIKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := s.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEmailDeliveryUnavailable, err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("%w: provider returned HTTP %d", ErrEmailDeliveryUnavailable, response.StatusCode)
	}
	return nil
}

func (s *Service) verifyEmailOTP(ctx context.Context, email, code string) (emailOTPVerification, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if len(code) != 6 {
		return emailOTPVerification{}, ErrInvalidOTP
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return emailOTPVerification{}, fmt.Errorf("begin email OTP verification: %w", err)
	}
	defer tx.Rollback()
	var challengeID, expectedHash []byte
	var purpose string
	var attempts int
	err = tx.QueryRowContext(ctx, `
		SELECT id, code_hash, purpose, attempt_count
		FROM email_otp_challenges
		WHERE email = ? AND consumed_at IS NULL AND expires_at > UTC_TIMESTAMP(6)
		ORDER BY created_at DESC LIMIT 1 FOR UPDATE`, email).Scan(&challengeID, &expectedHash, &purpose, &attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return emailOTPVerification{}, ErrInvalidOTP
	}
	if err != nil {
		return emailOTPVerification{}, fmt.Errorf("load email OTP challenge: %w", err)
	}
	if attempts >= 5 {
		return emailOTPVerification{}, ErrInvalidOTP
	}
	actualHash := s.emailOTPHash(challengeID, code)
	if !hmac.Equal(expectedHash, actualHash) {
		if _, err := tx.ExecContext(ctx, `UPDATE email_otp_challenges SET attempt_count = attempt_count + 1 WHERE id = ?`, challengeID); err != nil {
			return emailOTPVerification{}, fmt.Errorf("record incorrect email OTP: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return emailOTPVerification{}, fmt.Errorf("commit incorrect email OTP: %w", err)
		}
		return emailOTPVerification{}, ErrInvalidOTP
	}
	if _, err := tx.ExecContext(ctx, `UPDATE email_otp_challenges SET consumed_at = UTC_TIMESTAMP(6) WHERE id = ?`, challengeID); err != nil {
		return emailOTPVerification{}, fmt.Errorf("consume email OTP: %w", err)
	}

	switch purpose {
	case "registration":
		rawToken := make([]byte, 32)
		if _, err := rand.Read(rawToken); err != nil {
			return emailOTPVerification{}, fmt.Errorf("generate registration token: %w", err)
		}
		token := base64.RawURLEncoding.EncodeToString(rawToken)
		tokenHash := sha256.Sum256([]byte(token))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO registration_tokens (token_hash, email, expires_at)
			VALUES (?, ?, ?)`, tokenHash[:], email, s.now().UTC().Add(registrationTokenValidity)); err != nil {
			return emailOTPVerification{}, fmt.Errorf("store registration token: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return emailOTPVerification{}, fmt.Errorf("commit email verification: %w", err)
		}
		return emailOTPVerification{token: token, tokenEmail: email}, nil
	case "login":
		var accountID []byte
		err := tx.QueryRowContext(ctx, `
			SELECT account_id FROM account_contacts
			WHERE contact_type = 'email' AND value = ?`, email).Scan(&accountID)
		if errors.Is(err, sql.ErrNoRows) {
			return emailOTPVerification{}, ErrInvalidOTP
		}
		if err != nil {
			return emailOTPVerification{}, fmt.Errorf("load email OTP account: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE account_contacts SET verified_at = COALESCE(verified_at, UTC_TIMESTAMP(6))
			WHERE contact_type = 'email' AND value = ?`, email); err != nil {
			return emailOTPVerification{}, fmt.Errorf("verify account email: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return emailOTPVerification{}, fmt.Errorf("commit email login verification: %w", err)
		}
		principal, err := s.loadAccountPrincipal(ctx, accountID)
		if err != nil {
			return emailOTPVerification{}, fmt.Errorf("load email login account: %w", err)
		}
		principal, token, expiresAt, err := s.createSession(ctx, principal, "auth.email_otp_login")
		if err != nil {
			return emailOTPVerification{}, err
		}
		return emailOTPVerification{principal: &principal, session: token, expiresAt: expiresAt}, nil
	case "admin_bootstrap":
		if email != s.integrations.BootstrapAdminEmail {
			return emailOTPVerification{}, ErrInvalidOTP
		}
		var adminCount int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM account_roles WHERE role_code = 'admin'`).Scan(&adminCount); err != nil {
			return emailOTPVerification{}, fmt.Errorf("check administrator bootstrap: %w", err)
		}
		if adminCount != 0 {
			return emailOTPVerification{}, ErrInvalidOTP
		}
		accountID, err := randomID()
		if err != nil {
			return emailOTPVerification{}, fmt.Errorf("generate initial administrator ID: %w", err)
		}
		contactID, err := randomID()
		if err != nil {
			return emailOTPVerification{}, fmt.Errorf("generate initial administrator contact ID: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO accounts (id, email, name, account_type, status)
			VALUES (?, NULL, ?, 0, 'active')`, accountID, s.integrations.BootstrapAdminName); err != nil {
			return emailOTPVerification{}, fmt.Errorf("create initial administrator: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO account_contacts (id, account_id, contact_type, value, is_primary, verified_at)
			VALUES (?, ?, 'email', ?, TRUE, UTC_TIMESTAMP(6))`, contactID, accountID, email); err != nil {
			return emailOTPVerification{}, fmt.Errorf("verify initial administrator email: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO account_roles (account_id, role_code) VALUES (?, 'admin')`, accountID); err != nil {
			return emailOTPVerification{}, fmt.Errorf("grant initial administrator role: %w", err)
		}
		if err := insertAuditEventForResource(ctx, tx, nil, "account", accountID, "auth.bootstrap_admin", []byte(`{}`)); err != nil {
			return emailOTPVerification{}, fmt.Errorf("audit initial administrator: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return emailOTPVerification{}, fmt.Errorf("commit initial administrator: %w", err)
		}
		principal, err := s.loadAccountPrincipal(ctx, accountID)
		if err != nil {
			return emailOTPVerification{}, fmt.Errorf("load initial administrator: %w", err)
		}
		principal, token, expiresAt, err := s.createSession(ctx, principal, "auth.email_otp_admin_bootstrap")
		if err != nil {
			return emailOTPVerification{}, err
		}
		return emailOTPVerification{principal: &principal, session: token, expiresAt: expiresAt}, nil
	default:
		return emailOTPVerification{}, ErrInvalidOTP
	}
}

func (s *Service) completeRegistration(ctx context.Context, token string, profile registrationProfile) (Principal, string, time.Time, error) {
	if token == "" {
		return Principal{}, "", time.Time{}, ErrInvalidRegistration
	}
	tokenHash := sha256.Sum256([]byte(token))
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("begin alumni registration: %w", err)
	}
	defer tx.Rollback()
	var email string
	err = tx.QueryRowContext(ctx, `
		SELECT email FROM registration_tokens
		WHERE token_hash = ? AND consumed_at IS NULL AND expires_at > UTC_TIMESTAMP(6)
		FOR UPDATE`, tokenHash[:]).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return Principal{}, "", time.Time{}, ErrInvalidRegistration
	}
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("load registration token: %w", err)
	}
	accountID, err := randomID()
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("generate alumni account ID: %w", err)
	}
	contactID, err := randomID()
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("generate alumni contact ID: %w", err)
	}
	profileID, err := randomID()
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("generate alumni profile ID: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO accounts (id, email, name, account_type, status)
		VALUES (?, NULL, ?, 0, 'active')`, accountID, profile.fullName); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("create alumni account: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO account_contacts (id, account_id, contact_type, value, is_primary, verified_at)
		VALUES (?, ?, 'email', ?, TRUE, UTC_TIMESTAMP(6))`, contactID, accountID, email); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("store verified alumni email: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO account_roles (account_id, role_code) VALUES (?, 'member')`, accountID); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("assign alumni member role: %w", err)
	}
	if err := insertAlumniProfile(ctx, tx, profileID, accountID, profile); err != nil {
		return Principal{}, "", time.Time{}, err
	}
	discardPhoto, err := s.storePhoto(ctx, tx, accountID, profile.photo, profile.mimeType)
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("store alumni profile photo: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			discardPhoto()
		}
	}()
	if _, err := tx.ExecContext(ctx, `
		UPDATE registration_tokens SET consumed_at = UTC_TIMESTAMP(6)
		WHERE token_hash = ? AND consumed_at IS NULL`, tokenHash[:]); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("consume registration token: %w", err)
	}
	if err := insertAuditEventForResource(ctx, tx, nil, "account", accountID, "auth.registration", []byte(`{}`)); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("audit alumni registration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("commit alumni registration: %w", err)
	}
	committed = true
	principal, err := s.loadAccountPrincipal(ctx, accountID)
	if err != nil {
		return Principal{}, "", time.Time{}, fmt.Errorf("load new alumni account: %w", err)
	}
	principal, session, expiresAt, err := s.createSession(ctx, principal, "auth.registration")
	if err != nil {
		return Principal{}, "", time.Time{}, err
	}
	return principal, session, expiresAt, nil
}

func (s *Service) completeOAuthProfile(ctx context.Context, accountID string, profile registrationProfile) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin OAuth profile completion: %w", err)
	}
	defer tx.Rollback()
	var existing []byte
	err = tx.QueryRowContext(ctx, `SELECT id FROM alumni_profiles WHERE account_id = ? FOR UPDATE`, principalIDBytes(accountID)).Scan(&existing)
	if err == nil {
		return ErrProfileAlreadyComplete
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check OAuth profile completion: %w", err)
	}
	profileID, err := randomID()
	if err != nil {
		return fmt.Errorf("generate OAuth alumni profile ID: %w", err)
	}
	if err := insertAlumniProfile(ctx, tx, profileID, principalIDBytes(accountID), profile); err != nil {
		return err
	}
	discardPhoto, err := s.storePhoto(ctx, tx, principalIDBytes(accountID), profile.photo, profile.mimeType)
	if err != nil {
		return fmt.Errorf("store OAuth alumni profile photo: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			discardPhoto()
		}
	}()
	if _, err := tx.ExecContext(ctx, `UPDATE accounts SET name = ? WHERE id = ?`, profile.fullName, principalIDBytes(accountID)); err != nil {
		return fmt.Errorf("update OAuth alumni account name: %w", err)
	}
	if err := insertAuditEventForResource(ctx, tx, principalIDBytes(accountID), "alumni_profile", profileID, "alumni.profile_complete", []byte(`{}`)); err != nil {
		return fmt.Errorf("audit OAuth profile completion: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit OAuth profile completion: %w", err)
	}
	committed = true
	return nil
}

func (s *Service) profileCompletion(ctx context.Context, accountID []byte) (int, error) {
	var fullName, city string
	var batchYear int
	var hasPhoto, hasEmail, hasPhone, hasExternalIdentity bool
	err := s.db.QueryRowContext(ctx, `
		SELECT ap.full_name, ap.batch_year, ap.city,
			EXISTS(SELECT 1 FROM alumni_profile_photos p WHERE p.account_id = ap.account_id),
			EXISTS(SELECT 1 FROM account_contacts c WHERE c.account_id = ap.account_id
				AND c.contact_type = 'email' AND c.verified_at IS NOT NULL),
			EXISTS(SELECT 1 FROM account_contacts c WHERE c.account_id = ap.account_id
				AND c.contact_type = 'phone' AND c.verified_at IS NOT NULL),
			EXISTS(SELECT 1 FROM external_identities e WHERE e.account_id = ap.account_id)
		FROM alumni_profiles ap WHERE ap.account_id = ?`, accountID,
	).Scan(&fullName, &batchYear, &city, &hasPhoto, &hasEmail, &hasPhone, &hasExternalIdentity)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("query alumni profile completion: %w", err)
	}
	return calculateProfileCompletion(fullName, batchYear, city, hasPhoto, hasEmail, hasPhone, hasExternalIdentity), nil
}

func calculateProfileCompletion(fullName string, batchYear int, city string, hasPhoto, hasEmail, hasPhone, hasExternalIdentity bool) int {
	completion := 0
	if strings.TrimSpace(fullName) != "" {
		completion += 20
	}
	if batchYear > 0 {
		completion += 20
	}
	if strings.TrimSpace(city) != "" {
		completion += 20
	}
	if hasPhoto {
		completion += 20
	}
	if hasEmail || hasPhone || hasExternalIdentity {
		completion += 10
	}
	if hasEmail && hasPhone {
		completion += 10
	}
	return completion
}

func insertAlumniProfile(ctx context.Context, tx *sql.Tx, profileID, accountID []byte, profile registrationProfile) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO alumni_profiles
			(id, account_id, full_name, batch_year, city, country_code, city_code, verification_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'non_verified')`,
		profileID, accountID, profile.fullName, profile.batchYear,
		profile.city, profile.countryCode, nullableCityCode(profile.cityCode))
	if err != nil {
		return fmt.Errorf("store alumni profile: %w", err)
	}
	return nil
}

func nullableCityCode(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func validateRegistrationProfile(profile registrationProfile, currentYear int) error {
	profile.fullName = strings.TrimSpace(profile.fullName)
	if len([]rune(profile.fullName)) < 2 || len([]rune(profile.fullName)) > 160 {
		return ErrInvalidRegistration
	}
	if profile.batchYear < 1950 || profile.batchYear > currentYear {
		return ErrInvalidRegistration
	}
	profile.countryCode = strings.ToUpper(strings.TrimSpace(profile.countryCode))
	profile.cityCode = strings.TrimSpace(profile.cityCode)
	profile.city = strings.TrimSpace(profile.city)
	if !countryCodePattern.MatchString(profile.countryCode) || len([]rune(profile.city)) < 2 || len([]rune(profile.city)) > 120 {
		return ErrInvalidRegistration
	}
	if profile.countryCode == "ID" {
		if !domesticCityCodePattern.MatchString(profile.cityCode) {
			return ErrInvalidRegistration
		}
	} else if profile.cityCode != "" {
		return ErrInvalidRegistration
	}
	if len(profile.photo) == 0 || len(profile.photo) > maxProfilePhotoBytes {
		return ErrInvalidProfilePhoto
	}
	return nil
}

func parseRegistrationProfile(w http.ResponseWriter, r *http.Request, tokenRequired bool) (string, registrationProfile, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxProfilePhotoBytes+(64<<10))
	if err := r.ParseMultipartForm(maxProfilePhotoBytes + 64<<10); err != nil {
		return "", registrationProfile{}, ErrInvalidRegistration
	}
	token := strings.TrimSpace(r.FormValue("registration_token"))
	if tokenRequired && token == "" {
		return "", registrationProfile{}, ErrInvalidRegistration
	}
	batchYear, err := strconv.Atoi(strings.TrimSpace(r.FormValue("batch_year")))
	if err != nil {
		return "", registrationProfile{}, ErrInvalidRegistration
	}
	file, _, err := r.FormFile("photo")
	if err != nil {
		return "", registrationProfile{}, ErrInvalidProfilePhoto
	}
	defer file.Close()
	rawPhoto, err := io.ReadAll(io.LimitReader(file, maxProfilePhotoBytes+1))
	if err != nil || len(rawPhoto) > maxProfilePhotoBytes {
		return "", registrationProfile{}, ErrInvalidProfilePhoto
	}
	photo, mimeType, err := sanitizeProfilePhoto(rawPhoto)
	if err != nil {
		return "", registrationProfile{}, ErrInvalidProfilePhoto
	}
	profile := registrationProfile{
		fullName: r.FormValue("full_name"), batchYear: batchYear,
		countryCode: r.FormValue("country_code"), cityCode: r.FormValue("city_code"),
		city: r.FormValue("city"), photo: photo, mimeType: mimeType,
	}
	if err := validateRegistrationProfile(profile, time.Now().Year()); err != nil {
		return "", registrationProfile{}, err
	}
	profile.fullName = strings.TrimSpace(profile.fullName)
	profile.countryCode = strings.ToUpper(strings.TrimSpace(profile.countryCode))
	profile.cityCode = strings.TrimSpace(profile.cityCode)
	profile.city = strings.TrimSpace(profile.city)
	return token, profile, nil
}

func sanitizeProfilePhoto(raw []byte) ([]byte, string, error) {
	if len(raw) == 0 || len(raw) > maxProfilePhotoBytes {
		return nil, "", ErrInvalidProfilePhoto
	}
	decoded, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, "", ErrInvalidProfilePhoto
	}
	bounds := decoded.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width < 1 || height < 1 || width > 3000 || height > 3000 || int64(width)*int64(height) > 8_000_000 {
		return nil, "", ErrInvalidProfilePhoto
	}
	var output bytes.Buffer
	var contentType string
	switch format {
	case "jpeg":
		if err := jpeg.Encode(&output, decoded, &jpeg.Options{Quality: 85}); err != nil {
			return nil, "", ErrInvalidProfilePhoto
		}
		contentType = mime.TypeByExtension(".jpg")
	case "png":
		if err := png.Encode(&output, decoded); err != nil {
			return nil, "", ErrInvalidProfilePhoto
		}
		contentType = mime.TypeByExtension(".png")
	default:
		return nil, "", ErrInvalidProfilePhoto
	}
	if output.Len() == 0 || output.Len() > maxProfilePhotoBytes {
		return nil, "", ErrInvalidProfilePhoto
	}
	return output.Bytes(), contentType, nil
}

func (h *Handler) emailOTPStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !h.service.sameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	if !h.otpIPLimiter.Allow(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "too many OTP requests")
		return
	}
	var input struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if !h.otpEmailLimiter.Allow(email) {
		writeError(w, http.StatusTooManyRequests, "too many OTP requests")
		return
	}
	err := h.service.startEmailOTP(r.Context(), email)
	if errors.Is(err, ErrInvalidRegistration) {
		writeError(w, http.StatusBadRequest, "enter a valid email address")
		return
	}
	if errors.Is(err, ErrEmailDeliveryUnavailable) {
		h.reportFailure("email_otp_delivery", err)
		writeError(w, http.StatusServiceUnavailable, "email verification is temporarily unavailable")
		return
	}
	if err != nil {
		h.reportFailure("email_otp_start", err)
		writeError(w, http.StatusInternalServerError, "could not start email verification")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"message": "if this address can be used, a verification code has been sent",
		"content": nil,
	})
}

func (h *Handler) emailOTPVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !h.service.sameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	if !h.otpIPLimiter.Allow(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "too many OTP attempts")
		return
	}
	var input struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	result, err := h.service.verifyEmailOTP(r.Context(), input.Email, strings.TrimSpace(input.Code))
	if errors.Is(err, ErrInvalidOTP) {
		writeError(w, http.StatusBadRequest, "verification code is invalid or expired")
		return
	}
	if err != nil {
		h.reportFailure("email_otp_verify", err)
		writeError(w, http.StatusInternalServerError, "could not verify email address")
		return
	}
	if result.principal != nil {
		h.service.SetSessionCookie(w, result.session, result.expiresAt)
		writeJSON(w, http.StatusOK, map[string]any{
			"message": "email verified",
			"content": map[string]any{"user": result.principal, "expires_at": result.expiresAt},
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "email verified",
		"content": map[string]any{"registration_token": result.token, "email": result.tokenEmail},
	})
}

func (h *Handler) registerEmailMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !h.service.sameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	token, profile, err := parseRegistrationProfile(w, r, true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "profile data or photo is invalid")
		return
	}
	if !h.service.regencyCodeKnown(r.Context(), profile) {
		writeError(w, http.StatusBadRequest, "profile data or photo is invalid")
		return
	}
	principal, session, expiresAt, err := h.service.completeRegistration(r.Context(), token, profile)
	if errors.Is(err, ErrInvalidRegistration) {
		writeError(w, http.StatusBadRequest, "verification expired; request a new code")
		return
	}
	if err != nil {
		h.reportFailure("email_registration", err)
		writeError(w, http.StatusInternalServerError, "could not complete registration")
		return
	}
	h.service.SetSessionCookie(w, session, expiresAt)
	principal.ProfileCompletion = 90
	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "registration complete",
		"content": map[string]any{"user": principal, "expires_at": expiresAt, "profile_completion": principal.ProfileCompletion},
	})
}

func (h *Handler) completeOAuthProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	principal, ok := PrincipalFromContext(r.Context())
	if !ok || principal.Role != RoleOnboarding {
		writeError(w, http.StatusConflict, "profile onboarding is not required")
		return
	}
	_, profile, err := parseRegistrationProfile(w, r, false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "profile data or photo is invalid")
		return
	}
	if !h.service.regencyCodeKnown(r.Context(), profile) {
		writeError(w, http.StatusBadRequest, "profile data or photo is invalid")
		return
	}
	if err := h.service.completeOAuthProfile(r.Context(), principal.AccountID, profile); err != nil {
		if errors.Is(err, ErrProfileAlreadyComplete) {
			writeError(w, http.StatusConflict, "profile is already complete")
			return
		}
		h.reportFailure("oauth_profile_completion", err)
		writeError(w, http.StatusInternalServerError, "could not complete alumni profile")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "profile complete",
		"content": map[string]any{"profile_completion": 90},
	})
}

func registerEmailOTPRoutes(h *Handler, router chi.Router) {
	router.Post("/api/v1/auth/email-otp/start", h.emailOTPStart)
	router.Post("/api/v1/auth/email-otp/verify", h.emailOTPVerify)
	router.Post("/api/v1/auth/register", h.registerEmailMember)
	router.With(h.Authenticate).Post("/api/v1/auth/profile/complete", h.completeOAuthProfile)
}

var (
	ErrInvalidOTP               = errors.New("email OTP is invalid or expired")
	ErrInvalidRegistration      = errors.New("registration data is invalid")
	ErrInvalidProfilePhoto      = errors.New("profile photo is invalid")
	ErrEmailDeliveryUnavailable = errors.New("email delivery is unavailable")
	ErrProfileAlreadyComplete   = errors.New("profile is already complete")
)
