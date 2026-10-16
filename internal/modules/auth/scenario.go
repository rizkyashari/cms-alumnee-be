package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const scenarioBatchYear = 2012

type scenarioAccount struct {
	name string
	role string // member, officer or admin
}

// scenarioAccounts is the fixed cast for walking through the whole member
// lifecycle on a staging server: one Admin, two Pengurus and seven
// Non-verified classmates of the same batch (enough for the first-member
// confirmation plus a candidate with five attesters).
var scenarioAccounts = []scenarioAccount{
	{"Admin Uji", "admin"},
	{"Pengurus Satu", "officer"},
	{"Pengurus Dua", "officer"},
	{"Alumni Satu", "member"},
	{"Alumni Dua", "member"},
	{"Alumni Tiga", "member"},
	{"Alumni Empat", "member"},
	{"Alumni Lima", "member"},
	{"Alumni Enam", "member"},
	{"Alumni Tujuh", "member"},
}

// ProvisionScenarioAccounts creates the staging scenario accounts, all sharing
// one password; usernames come from configuration in cast order. It must run after ProvisionStagingAccounts, which disables
// every password credential that is not configured. An empty password leaves
// the accounts disabled. Existing members keep their verification progress
// unless reset is true, which returns them to a clean Non-verified state.
func (s *Service) ProvisionScenarioAccounts(ctx context.Context, password string, usernames []string, reset bool) error {
	if password == "" {
		return nil
	}
	if len(password) < 12 {
		return errors.New("scenario account password must contain at least 12 characters")
	}
	if len(usernames) != len(scenarioAccounts) {
		return fmt.Errorf("expected %d scenario usernames, got %d", len(scenarioAccounts), len(usernames))
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash scenario account password: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin scenario provisioning: %w", err)
	}
	defer tx.Rollback()

	for index, account := range scenarioAccounts {
		username := usernames[index]
		var accountID []byte
		err := tx.QueryRowContext(ctx, `SELECT account_id FROM password_credentials WHERE username = ?`, username).Scan(&accountID)
		if errors.Is(err, sql.ErrNoRows) {
			if accountID, err = randomID(); err != nil {
				return fmt.Errorf("generate scenario account ID: %w", err)
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO accounts (id, email, name, account_type, status) VALUES (?, NULL, ?, 0, 'active')`, accountID, account.name); err != nil {
				return fmt.Errorf("create scenario account: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("find scenario account: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE accounts SET name = ?, status = 'active' WHERE id = ?`, account.name, accountID); err != nil {
			return fmt.Errorf("update scenario account: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO password_credentials (username, account_id, password_hash, is_enabled)
			VALUES (?, ?, ?, TRUE)
			ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), is_enabled = TRUE`,
			username, accountID, hash); err != nil {
			return fmt.Errorf("store scenario credentials: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO account_roles (account_id, role_code) VALUES (?, ?)
			ON DUPLICATE KEY UPDATE role_code = role_code`, accountID, account.role); err != nil {
			return fmt.Errorf("assign scenario role: %w", err)
		}
		if account.role != "member" {
			continue
		}
		profileID, err := randomID()
		if err != nil {
			return fmt.Errorf("generate scenario profile ID: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO alumni_profiles (id, account_id, full_name, batch_year, city, verification_status)
			VALUES (?, ?, ?, ?, 'Yogyakarta', 'non_verified')
			ON DUPLICATE KEY UPDATE full_name = VALUES(full_name), batch_year = VALUES(batch_year)`,
			profileID, accountID, account.name, scenarioBatchYear); err != nil {
			return fmt.Errorf("provision scenario profile: %w", err)
		}
		if reset {
			if _, err := tx.ExecContext(ctx, `
				DELETE FROM verification_requests
				WHERE profile_id = (SELECT id FROM alumni_profiles WHERE account_id = ?)`, accountID); err != nil {
				return fmt.Errorf("reset scenario verification request: %w", err)
			}
			if _, err := tx.ExecContext(ctx, `
				DELETE FROM alumni_attestations
				WHERE profile_id = (SELECT id FROM alumni_profiles WHERE account_id = ?)
					OR attester_id = ?`, accountID, accountID); err != nil {
				return fmt.Errorf("reset scenario attestations: %w", err)
			}
			if _, err := tx.ExecContext(ctx, `UPDATE alumni_profiles SET verification_status = 'non_verified' WHERE account_id = ?`, accountID); err != nil {
				return fmt.Errorf("reset scenario verification status: %w", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit scenario provisioning: %w", err)
	}
	s.passwordLoginEnabled = true
	return nil
}
