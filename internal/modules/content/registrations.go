package content

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	RegistrationNone       = "none"
	RegistrationRegistered = "registered"
	RegistrationCancelled  = "cancelled"
	RegistrationAttended   = "attended"
)

var (
	ErrProfileRequired      = errors.New("alumni profile is required to register")
	ErrEventFinished        = errors.New("event has already finished")
	ErrEventNotStarted      = errors.New("event has not started yet")
	ErrRegistrationNotFound = errors.New("registration not found")
	ErrNotOrganizer         = errors.New("only the event author or an Admin may manage attendance")
)

// MyRegistration pairs a registration with its (published) event.
type MyRegistration struct {
	Status string `json:"status"`
	Event  Event  `json:"event"`
}

type Attendee struct {
	ProfileID    string `json:"profile_id"`
	Name         string `json:"name"`
	BatchYear    int    `json:"batch_year"`
	Status       string `json:"status"`
	RegisteredAt string `json:"registered_at"`
}

func (s *Service) profileID(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, accountID []byte) ([]byte, error) {
	var id []byte
	err := q.QueryRowContext(ctx, `SELECT id FROM alumni_profiles WHERE account_id = ?`, accountID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProfileRequired
	}
	if err != nil {
		return nil, fmt.Errorf("load alumni profile: %w", err)
	}
	return id, nil
}

// RegistrationStatus returns the caller's registration for an event
// (none | registered | cancelled | attended). The event must be published.
func (s *Service) RegistrationStatus(ctx context.Context, eventID, accountID []byte) (string, error) {
	if _, err := s.GetPublished(ctx, eventID); err != nil {
		return "", err
	}
	profile, err := s.profileID(ctx, s.db, accountID)
	if errors.Is(err, ErrProfileRequired) {
		return RegistrationNone, nil
	}
	if err != nil {
		return "", err
	}
	var status string
	err = s.db.QueryRowContext(ctx, `SELECT status FROM event_registrations WHERE event_id = ? AND profile_id = ?`, eventID, profile).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return RegistrationNone, nil
	}
	if err != nil {
		return "", fmt.Errorf("load registration: %w", err)
	}
	return status, nil
}

// Register signs the caller up for a published, unfinished event. It is
// idempotent for registered/attended members and re-opens a cancelled one.
func (s *Service) Register(ctx context.Context, eventID, accountID []byte) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin register: %w", err)
	}
	defer tx.Rollback()
	profile, err := s.profileID(ctx, tx, accountID)
	if err != nil {
		return "", err
	}
	_, status, err := lockEvent(ctx, tx, eventID)
	if err != nil {
		return "", err
	}
	if status != StatusPublished {
		return "", ErrEventNotFound
	}
	var startsAt time.Time
	var endsAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT starts_at, ends_at FROM alumni_events WHERE id = ?`, eventID).Scan(&startsAt, &endsAt); err != nil {
		return "", fmt.Errorf("load event window: %w", err)
	}
	end := startsAt
	if endsAt.Valid {
		end = endsAt.Time
	}
	var current string
	err = tx.QueryRowContext(ctx, `SELECT status FROM event_registrations WHERE event_id = ? AND profile_id = ? FOR UPDATE`, eventID, profile).Scan(&current)
	switch {
	case err == nil && (current == RegistrationRegistered || current == RegistrationAttended):
		return current, nil
	case err != nil && !errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf("load registration: %w", err)
	}
	if !end.After(s.now()) {
		return "", ErrEventFinished
	}
	if errors.Is(err, sql.ErrNoRows) {
		id, err := randomID()
		if err != nil {
			return "", fmt.Errorf("generate registration ID: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO event_registrations (id, event_id, profile_id) VALUES (?, ?, ?)`, id, eventID, profile); err != nil {
			return "", fmt.Errorf("insert registration: %w", err)
		}
	} else if _, err := tx.ExecContext(ctx, `UPDATE event_registrations SET status = 'registered' WHERE event_id = ? AND profile_id = ?`, eventID, profile); err != nil {
		return "", fmt.Errorf("reopen registration: %w", err)
	}
	if err := audit(ctx, tx, accountID, eventID, "event.register", nil); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit register: %w", err)
	}
	return RegistrationRegistered, nil
}

// Unregister cancels a registered (not yet attended) registration.
func (s *Service) Unregister(ctx context.Context, eventID, accountID []byte) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin unregister: %w", err)
	}
	defer tx.Rollback()
	profile, err := s.profileID(ctx, tx, accountID)
	if errors.Is(err, ErrProfileRequired) {
		return ErrRegistrationNotFound
	}
	if err != nil {
		return err
	}
	var current string
	err = tx.QueryRowContext(ctx, `SELECT status FROM event_registrations WHERE event_id = ? AND profile_id = ? FOR UPDATE`, eventID, profile).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrRegistrationNotFound
	}
	if err != nil {
		return fmt.Errorf("load registration: %w", err)
	}
	if current == RegistrationCancelled {
		return nil
	}
	if current != RegistrationRegistered {
		return ErrInvalidState
	}
	if _, err := tx.ExecContext(ctx, `UPDATE event_registrations SET status = 'cancelled' WHERE event_id = ? AND profile_id = ?`, eventID, profile); err != nil {
		return fmt.Errorf("cancel registration: %w", err)
	}
	if err := audit(ctx, tx, accountID, eventID, "event.unregister", nil); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit unregister: %w", err)
	}
	return nil
}

// ListMine returns the caller's non-cancelled registrations for published
// events, upcoming first then most recent.
func (s *Service) ListMine(ctx context.Context, accountID []byte) ([]MyRegistration, error) {
	items := []MyRegistration{}
	profile, err := s.profileID(ctx, s.db, accountID)
	if errors.Is(err, ErrProfileRequired) {
		return items, nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT er.status, e.id FROM event_registrations er
		JOIN alumni_events e ON e.id = er.event_id AND e.status = 'published'
		WHERE er.profile_id = ? AND er.status <> 'cancelled'
		ORDER BY (COALESCE(e.ends_at, e.starts_at) >= ?) DESC, e.starts_at ASC LIMIT 100`, profile, s.now().UTC())
	if err != nil {
		return nil, fmt.Errorf("list registrations: %w", err)
	}
	type pair struct {
		status string
		id     []byte
	}
	var pairs []pair
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.status, &p.id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan registration: %w", err)
		}
		pairs = append(pairs, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for _, p := range pairs {
		event, err := s.load(ctx, p.id)
		if err != nil {
			return nil, err
		}
		items = append(items, MyRegistration{Status: p.status, Event: event})
	}
	return items, nil
}

// authorizeOrganizer allows the event's author or a moderator.
func (s *Service) authorizeOrganizer(ctx context.Context, eventID, actor []byte, moderator bool) error {
	var author []byte
	err := s.db.QueryRowContext(ctx, `SELECT author_id FROM alumni_events WHERE id = ?`, eventID).Scan(&author)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrEventNotFound
	}
	if err != nil {
		return fmt.Errorf("load event author: %w", err)
	}
	if !moderator && !isOwner(author, actor) {
		return ErrNotOrganizer
	}
	return nil
}

// ListAttendees returns active registrations for an event to its author or an Admin.
func (s *Service) ListAttendees(ctx context.Context, eventID, actor []byte, moderator bool) ([]Attendee, error) {
	if err := s.authorizeOrganizer(ctx, eventID, actor, moderator); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.id, p.full_name, p.batch_year, er.status, er.created_at
		FROM event_registrations er JOIN alumni_profiles p ON p.id = er.profile_id
		WHERE er.event_id = ? AND er.status <> 'cancelled' ORDER BY er.created_at, er.id`, eventID)
	if err != nil {
		return nil, fmt.Errorf("list attendees: %w", err)
	}
	defer rows.Close()
	items := []Attendee{}
	for rows.Next() {
		var a Attendee
		var id []byte
		var at time.Time
		if err := rows.Scan(&id, &a.Name, &a.BatchYear, &a.Status, &at); err != nil {
			return nil, fmt.Errorf("scan attendee: %w", err)
		}
		a.ProfileID = formatID(id)
		a.RegisteredAt = at.UTC().Format(time.RFC3339)
		items = append(items, a)
	}
	return items, rows.Err()
}

// MarkAttended records attendance once the event has started.
func (s *Service) MarkAttended(ctx context.Context, eventID, profileID, actor []byte, moderator bool) error {
	if err := s.authorizeOrganizer(ctx, eventID, actor, moderator); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin attendance: %w", err)
	}
	defer tx.Rollback()
	var startsAt time.Time
	if err := tx.QueryRowContext(ctx, `SELECT starts_at FROM alumni_events WHERE id = ? FOR UPDATE`, eventID).Scan(&startsAt); err != nil {
		return fmt.Errorf("lock event: %w", err)
	}
	if startsAt.After(s.now()) {
		return ErrEventNotStarted
	}
	var current string
	err = tx.QueryRowContext(ctx, `SELECT status FROM event_registrations WHERE event_id = ? AND profile_id = ? FOR UPDATE`, eventID, profileID).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrRegistrationNotFound
	}
	if err != nil {
		return fmt.Errorf("load registration: %w", err)
	}
	if current == RegistrationAttended {
		return nil
	}
	if current != RegistrationRegistered {
		return ErrInvalidState
	}
	if _, err := tx.ExecContext(ctx, `UPDATE event_registrations SET status = 'attended' WHERE event_id = ? AND profile_id = ?`, eventID, profileID); err != nil {
		return fmt.Errorf("mark attended: %w", err)
	}
	if err := audit(ctx, tx, actor, eventID, "event.attend", map[string]string{"profile_id": formatID(profileID)}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit attendance: %w", err)
	}
	return nil
}
