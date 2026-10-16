package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type managedRole struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      Role   `json:"role"`
	GrantedAt string `json:"granted_at"`
}

func (s *Service) listManagedRoles(ctx context.Context) ([]managedRole, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.value, a.name, ar.role_code, ar.granted_at
		FROM account_roles ar
		JOIN accounts a ON a.id = ar.account_id
		JOIN account_contacts c ON c.account_id = a.id
			AND c.contact_type = 'email' AND c.is_primary = TRUE AND c.verified_at IS NOT NULL
		WHERE ar.role_code IN ('admin', 'officer') AND a.status = 'active'
		ORDER BY ar.role_code, a.name, c.value`)
	if err != nil {
		return nil, fmt.Errorf("query managed roles: %w", err)
	}
	defer rows.Close()
	roles := make([]managedRole, 0)
	for rows.Next() {
		var item managedRole
		var role string
		var grantedAt time.Time
		if err := rows.Scan(&item.Email, &item.Name, &role, &grantedAt); err != nil {
			return nil, fmt.Errorf("scan managed role: %w", err)
		}
		item.Role = Role(role)
		item.GrantedAt = grantedAt.UTC().Format(time.RFC3339)
		roles = append(roles, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate managed roles: %w", err)
	}
	return roles, nil
}

func (s *Service) changeManagedRole(ctx context.Context, actorID, rawEmail string, role Role, grant bool) error {
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 || (role != RoleAdmin && role != RoleOfficer) {
		return ErrInvalidManagedRole
	}
	actorBytes := principalIDBytes(actorID)
	if len(actorBytes) != 16 {
		return ErrInvalidManagedRole
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin managed role change: %w", err)
	}
	defer tx.Rollback()
	var activeAdmins int
	adminRows, err := tx.QueryContext(ctx, `
		SELECT ar.account_id FROM account_roles ar
		JOIN accounts a ON a.id = ar.account_id
		WHERE ar.role_code = 'admin' AND a.status = 'active'
		ORDER BY ar.account_id FOR UPDATE`)
	if err != nil {
		return fmt.Errorf("lock active administrator roles: %w", err)
	}
	for adminRows.Next() {
		var ignored []byte
		if err := adminRows.Scan(&ignored); err != nil {
			adminRows.Close()
			return fmt.Errorf("scan active administrator role: %w", err)
		}
		activeAdmins++
	}
	if err := adminRows.Err(); err != nil {
		adminRows.Close()
		return fmt.Errorf("iterate active administrator roles: %w", err)
	}
	if err := adminRows.Close(); err != nil {
		return fmt.Errorf("close active administrator roles: %w", err)
	}
	var targetID []byte
	err = tx.QueryRowContext(ctx, `
		SELECT a.id FROM accounts a
		JOIN account_contacts c ON c.account_id = a.id
		WHERE c.contact_type = 'email' AND c.value = ?
			AND c.verified_at IS NOT NULL AND a.status = 'active'
		FOR UPDATE`, email).Scan(&targetID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrManagedAccountNotFound
	}
	if err != nil {
		return fmt.Errorf("find managed account: %w", err)
	}

	action := "admin.role_grant"
	if grant {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO account_roles (account_id, role_code, granted_by)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE granted_by = VALUES(granted_by), granted_at = UTC_TIMESTAMP(6)`,
			targetID, string(role), actorBytes)
	} else {
		action = "admin.role_revoke"
		var assignmentExists bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM account_roles WHERE account_id = ? AND role_code = ?
			)`, targetID, string(role)).Scan(&assignmentExists); err != nil {
			return fmt.Errorf("check staff role assignment: %w", err)
		}
		if !assignmentExists {
			return ErrRoleAssignmentNotFound
		}
		if role == RoleAdmin && activeAdmins <= 1 {
			return ErrLastAdministrator
		}
		var result sql.Result
		result, err = tx.ExecContext(ctx, `DELETE FROM account_roles WHERE account_id = ? AND role_code = ?`, targetID, string(role))
		if err == nil {
			affected, affectedErr := result.RowsAffected()
			if affectedErr != nil {
				return fmt.Errorf("check role revocation result: %w", affectedErr)
			}
			if affected == 0 {
				return ErrRoleAssignmentNotFound
			}
		}
	}
	if err != nil {
		return fmt.Errorf("change managed role: %w", err)
	}
	metadata, err := json.Marshal(map[string]string{"role": string(role)})
	if err != nil {
		return fmt.Errorf("encode role audit metadata: %w", err)
	}
	auditID, err := randomID()
	if err != nil {
		return fmt.Errorf("generate role audit ID: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO audit_events (id, actor_id, action, resource_type, resource_id, metadata)
		VALUES (?, ?, ?, 'account_role', ?, ?)`,
		auditID, actorBytes, action, targetID, metadata); err != nil {
		return fmt.Errorf("audit managed role change: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit managed role change: %w", err)
	}
	return nil
}

func (h *Handler) managedRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.listManagedRoles(r.Context())
	if err != nil {
		h.reportFailure("list_managed_roles", err)
		writeError(w, http.StatusInternalServerError, "could not load staff accounts")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "ok", "content": map[string]any{"items": roles},
	})
}

func (h *Handler) changeManagedRole(w http.ResponseWriter, r *http.Request) {
	if !h.service.sameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	operation := chi.URLParam(r, "operation")
	if operation != "grant" && operation != "revoke" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	var input struct {
		Email string `json:"email"`
		Role  Role   `json:"role"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	grant := operation == "grant"
	err := h.service.changeManagedRole(r.Context(), principal.AccountID, input.Email, input.Role, grant)
	switch {
	case errors.Is(err, ErrInvalidManagedRole):
		writeError(w, http.StatusBadRequest, "enter a valid email and staff role")
		return
	case errors.Is(err, ErrManagedAccountNotFound):
		writeError(w, http.StatusNotFound, "no active account with that verified email was found")
		return
	case errors.Is(err, ErrLastAdministrator):
		writeError(w, http.StatusConflict, "the last active administrator cannot be removed")
		return
	case errors.Is(err, ErrRoleAssignmentNotFound):
		writeError(w, http.StatusNotFound, "that staff role is not assigned to the account")
		return
	case err != nil:
		h.reportFailure("change_managed_role", err)
		writeError(w, http.StatusInternalServerError, "could not change staff role")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "staff role updated", "content": nil})
}

var (
	ErrInvalidManagedRole     = errors.New("invalid managed role")
	ErrManagedAccountNotFound = errors.New("managed account not found")
	ErrLastAdministrator      = errors.New("cannot remove the last administrator")
	ErrRoleAssignmentNotFound = errors.New("staff role assignment not found")
)
