// Package content owns community content. Agenda (alumni events) follows the
// review workflow draft -> pending_review -> published | rejected -> archived;
// only the backend moves an event between states and every transition is
// audited.
package content

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	StatusDraft         = "draft"
	StatusPendingReview = "pending_review"
	StatusPublished     = "published"
	StatusRejected      = "rejected"
	StatusArchived      = "archived"

	maxTitleRunes       = 200
	maxVenueRunes       = 200
	maxDescriptionRunes = 5000
	maxNoteRunes        = 500
	defaultPageSize     = 20
	maxPageSize         = 50
)

var eventCategories = map[string]struct{}{
	"reuni": {}, "karier": {}, "seminar": {}, "sosial": {}, "lainnya": {},
}

var (
	ErrEventNotFound  = errors.New("event not found")
	ErrInvalidState   = errors.New("event is not in a state that allows this action")
	ErrInvalidInput   = errors.New("invalid event input")
	ErrStartsInPast   = errors.New("event must start in the future")
	ErrNoteRequired   = errors.New("a note is required when rejecting an event")
	ErrUnknownDecison = errors.New("unknown decision")
)

// ValidationError carries a user-presentable reason; it unwraps to ErrInvalidInput.
type ValidationError struct{ Reason string }

func (e ValidationError) Error() string { return e.Reason }
func (e ValidationError) Unwrap() error { return ErrInvalidInput }

type Event struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Category      string  `json:"category"`
	Venue         string  `json:"venue"`
	Description   string  `json:"description"`
	StartsAt      string  `json:"starts_at"`
	EndsAt        *string `json:"ends_at"`
	Status        string  `json:"status"`
	AuthorName    string  `json:"author_name,omitempty"`
	ReviewNote    string  `json:"review_note,omitempty"`
	PublishedAt   *string `json:"published_at"`
	CreatedAt     string  `json:"created_at"`
	CommentCount  int     `json:"comment_count"`
	AttendeeCount int     `json:"attendee_count"`
}

type EventInput struct {
	Title       string  `json:"title"`
	Category    string  `json:"category"`
	Venue       string  `json:"venue"`
	Description string  `json:"description"`
	StartsAt    string  `json:"starts_at"`
	EndsAt      *string `json:"ends_at"`
}

type normalizedEvent struct {
	title, category, venue, description string
	startsAt                            time.Time
	endsAt                              *time.Time
}

type ListFilter struct {
	Query    string
	Category string
	Scope    string // upcoming (default) | past | all
	Status   string // admin and own listings only
	Limit    int
	Offset   int
}

type EventPage struct {
	Items  []Event `json:"items"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

type Service struct {
	db           *sql.DB
	now          func() time.Time
	commentLimit commentLimiter
}

func NewService(db *sql.DB) *Service { return &Service{db: db, now: time.Now} }

func validateInput(in EventInput) (normalizedEvent, error) {
	out := normalizedEvent{
		title:       strings.TrimSpace(in.Title),
		category:    strings.ToLower(strings.TrimSpace(in.Category)),
		venue:       strings.TrimSpace(in.Venue),
		description: strings.TrimSpace(in.Description),
	}
	if n := utf8.RuneCountInString(out.title); n < 3 || n > maxTitleRunes {
		return out, ValidationError{"title must be 3-200 characters"}
	}
	if _, ok := eventCategories[out.category]; !ok {
		return out, ValidationError{"category must be one of reuni, karier, seminar, sosial, lainnya"}
	}
	if utf8.RuneCountInString(out.venue) > maxVenueRunes {
		return out, ValidationError{"venue must be at most 200 characters"}
	}
	if out.description == "" || utf8.RuneCountInString(out.description) > maxDescriptionRunes {
		return out, ValidationError{"description is required and must be at most 5000 characters"}
	}
	starts, err := time.Parse(time.RFC3339, strings.TrimSpace(in.StartsAt))
	if err != nil {
		return out, ValidationError{"starts_at must be an RFC 3339 timestamp"}
	}
	out.startsAt = starts.UTC()
	if in.EndsAt != nil && strings.TrimSpace(*in.EndsAt) != "" {
		ends, err := time.Parse(time.RFC3339, strings.TrimSpace(*in.EndsAt))
		if err != nil {
			return out, ValidationError{"ends_at must be an RFC 3339 timestamp"}
		}
		if !ends.After(starts) {
			return out, ValidationError{"ends_at must be after starts_at"}
		}
		utc := ends.UTC()
		out.endsAt = &utc
	}
	return out, nil
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

func parseID(id string) []byte {
	result, err := hex.DecodeString(strings.ReplaceAll(id, "-", ""))
	if err != nil || len(result) != 16 {
		return nil
	}
	return result
}

func audit(ctx context.Context, tx *sql.Tx, actor, resource []byte, action string, meta map[string]string, resourceType ...string) error {
	kind := "event"
	if len(resourceType) > 0 {
		kind = resourceType[0]
	}
	id, err := randomID()
	if err != nil {
		return fmt.Errorf("generate audit ID: %w", err)
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("encode audit metadata: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO audit_events (id, actor_id, action, resource_type, resource_id, metadata)
		VALUES (?, ?, ?, ?, ?, ?)`, id, actor, action, kind, resource, string(payload))
	if err != nil {
		return fmt.Errorf("audit %s: %w", action, err)
	}
	return nil
}

func (s *Service) CreateDraft(ctx context.Context, authorID []byte, in EventInput) (Event, error) {
	n, err := validateInput(in)
	if err != nil {
		return Event{}, err
	}
	id, err := randomID()
	if err != nil {
		return Event{}, fmt.Errorf("generate event ID: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Event{}, fmt.Errorf("begin create event: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO alumni_events (id, author_id, title, category, venue, description, starts_at, ends_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'draft')`,
		id, authorID, n.title, n.category, n.venue, n.description, n.startsAt, n.endsAt); err != nil {
		return Event{}, fmt.Errorf("insert event: %w", err)
	}
	if err := audit(ctx, tx, authorID, id, "event.create", nil); err != nil {
		return Event{}, err
	}
	if err := tx.Commit(); err != nil {
		return Event{}, fmt.Errorf("commit create event: %w", err)
	}
	return s.load(ctx, id)
}

// lockEvent returns the event's author and status inside tx, locking the row.
func lockEvent(ctx context.Context, tx *sql.Tx, id []byte) (author []byte, status string, err error) {
	err = tx.QueryRowContext(ctx, `SELECT author_id, status FROM alumni_events WHERE id = ? FOR UPDATE`, id).Scan(&author, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", ErrEventNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("lock event: %w", err)
	}
	return author, status, nil
}

func isOwner(author, actor []byte) bool { return author != nil && string(author) == string(actor) }

// UpdateOwn edits an event the caller authored while it is a draft or was
// rejected; an edited rejected event returns to draft for resubmission.
func (s *Service) UpdateOwn(ctx context.Context, actor, id []byte, in EventInput) (Event, error) {
	n, err := validateInput(in)
	if err != nil {
		return Event{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Event{}, fmt.Errorf("begin update event: %w", err)
	}
	defer tx.Rollback()
	author, status, err := lockEvent(ctx, tx, id)
	if err != nil {
		return Event{}, err
	}
	if !isOwner(author, actor) {
		return Event{}, ErrEventNotFound
	}
	if status != StatusDraft && status != StatusRejected {
		return Event{}, ErrInvalidState
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE alumni_events SET title = ?, category = ?, venue = ?, description = ?, starts_at = ?, ends_at = ?,
			status = 'draft', review_note = NULL WHERE id = ?`,
		n.title, n.category, n.venue, n.description, n.startsAt, n.endsAt, id); err != nil {
		return Event{}, fmt.Errorf("update event: %w", err)
	}
	if err := audit(ctx, tx, actor, id, "event.update", nil); err != nil {
		return Event{}, err
	}
	if err := tx.Commit(); err != nil {
		return Event{}, fmt.Errorf("commit update event: %w", err)
	}
	return s.load(ctx, id)
}

func (s *Service) Submit(ctx context.Context, actor, id []byte) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin submit event: %w", err)
	}
	defer tx.Rollback()
	author, status, err := lockEvent(ctx, tx, id)
	if err != nil {
		return err
	}
	if !isOwner(author, actor) {
		return ErrEventNotFound
	}
	if status != StatusDraft {
		return ErrInvalidState
	}
	var startsAt time.Time
	if err := tx.QueryRowContext(ctx, `SELECT starts_at FROM alumni_events WHERE id = ?`, id).Scan(&startsAt); err != nil {
		return fmt.Errorf("load event start: %w", err)
	}
	if !startsAt.After(s.now()) {
		return ErrStartsInPast
	}
	if _, err := tx.ExecContext(ctx, `UPDATE alumni_events SET status = 'pending_review' WHERE id = ?`, id); err != nil {
		return fmt.Errorf("submit event: %w", err)
	}
	if err := audit(ctx, tx, actor, id, "event.submit", nil); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit submit event: %w", err)
	}
	return nil
}

// DeleteOwn removes an unpublished event the caller authored.
func (s *Service) DeleteOwn(ctx context.Context, actor, id []byte) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete event: %w", err)
	}
	defer tx.Rollback()
	author, status, err := lockEvent(ctx, tx, id)
	if err != nil {
		return err
	}
	if !isOwner(author, actor) {
		return ErrEventNotFound
	}
	if status != StatusDraft && status != StatusRejected {
		return ErrInvalidState
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM alumni_events WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	if err := audit(ctx, tx, actor, id, "event.delete", nil); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete event: %w", err)
	}
	return nil
}

// Decide applies an admin decision: publish and reject act on pending_review
// events, archive acts on published ones. Rejection requires a note.
func (s *Service) Decide(ctx context.Context, actor, id []byte, decision, note string) (Event, error) {
	note = strings.TrimSpace(note)
	if utf8.RuneCountInString(note) > maxNoteRunes {
		return Event{}, ValidationError{"note must be at most 500 characters"}
	}
	var from, to string
	switch decision {
	case "publish":
		from, to = StatusPendingReview, StatusPublished
	case "reject":
		from, to = StatusPendingReview, StatusRejected
		if note == "" {
			return Event{}, ErrNoteRequired
		}
	case "archive":
		from, to = StatusPublished, StatusArchived
	default:
		return Event{}, ErrUnknownDecison
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Event{}, fmt.Errorf("begin event decision: %w", err)
	}
	defer tx.Rollback()
	_, status, err := lockEvent(ctx, tx, id)
	if err != nil {
		return Event{}, err
	}
	if status != from {
		return Event{}, ErrInvalidState
	}
	var noteValue any
	if note != "" {
		noteValue = note
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE alumni_events SET status = ?, review_note = ?, reviewed_by = ?, reviewed_at = CURRENT_TIMESTAMP(6),
			published_at = IF(? = 'published', CURRENT_TIMESTAMP(6), published_at) WHERE id = ?`,
		to, noteValue, actor, to, id); err != nil {
		return Event{}, fmt.Errorf("decide event: %w", err)
	}
	if err := audit(ctx, tx, actor, id, "event."+decision, map[string]string{"from": from, "to": to}); err != nil {
		return Event{}, err
	}
	if err := tx.Commit(); err != nil {
		return Event{}, fmt.Errorf("commit event decision: %w", err)
	}
	return s.load(ctx, id)
}

const eventColumns = `
	SELECT e.id, e.title, e.category, e.venue, e.description, e.starts_at, e.ends_at, e.status,
		COALESCE(p.full_name, a.name, ''), COALESCE(e.review_note, ''), e.published_at, e.created_at,
		(SELECT COUNT(*) FROM event_comments c WHERE c.event_id = e.id),
		(SELECT COUNT(*) FROM event_registrations er WHERE er.event_id = e.id AND er.status <> 'cancelled')
	FROM alumni_events e
	LEFT JOIN accounts a ON a.id = e.author_id
	LEFT JOIN alumni_profiles p ON p.account_id = e.author_id`

type rowScanner interface{ Scan(...any) error }

func scanEvent(row rowScanner) (Event, error) {
	var (
		e                   Event
		id                  []byte
		startsAt, createdAt time.Time
		endsAt, publishedAt sql.NullTime
	)
	if err := row.Scan(&id, &e.Title, &e.Category, &e.Venue, &e.Description, &startsAt, &endsAt, &e.Status,
		&e.AuthorName, &e.ReviewNote, &publishedAt, &createdAt, &e.CommentCount, &e.AttendeeCount); err != nil {
		return Event{}, err
	}
	e.ID = formatID(id)
	e.StartsAt = startsAt.UTC().Format(time.RFC3339)
	e.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	if endsAt.Valid {
		v := endsAt.Time.UTC().Format(time.RFC3339)
		e.EndsAt = &v
	}
	if publishedAt.Valid {
		v := publishedAt.Time.UTC().Format(time.RFC3339)
		e.PublishedAt = &v
	}
	return e, nil
}

func (s *Service) load(ctx context.Context, id []byte) (Event, error) {
	e, err := scanEvent(s.db.QueryRowContext(ctx, eventColumns+` WHERE e.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, ErrEventNotFound
	}
	if err != nil {
		return Event{}, fmt.Errorf("load event: %w", err)
	}
	return e, nil
}

// GetPublished returns an event only when it is published.
func (s *Service) GetPublished(ctx context.Context, id []byte) (Event, error) {
	e, err := s.load(ctx, id)
	if err != nil {
		return Event{}, err
	}
	if e.Status != StatusPublished {
		return Event{}, ErrEventNotFound
	}
	return e, nil
}

func escapeLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}

// List returns one page of events. forcedStatus limits public listings to
// published events; authorID, when set, limits to that author's events.
func (s *Service) List(ctx context.Context, f ListFilter, publicOnly bool, authorID []byte) (EventPage, error) {
	if f.Limit <= 0 {
		f.Limit = defaultPageSize
	}
	if f.Limit > maxPageSize {
		f.Limit = maxPageSize
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	var where []string
	var args []any
	order := "e.starts_at DESC"
	switch {
	case publicOnly:
		where = append(where, "e.status = 'published'")
	case f.Status != "":
		where = append(where, "e.status = ?")
		args = append(args, f.Status)
	}
	if authorID != nil {
		where = append(where, "e.author_id = ?")
		args = append(args, authorID)
	}
	if f.Category != "" {
		where = append(where, "e.category = ?")
		args = append(args, strings.ToLower(f.Category))
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		like := "%" + escapeLike(q) + "%"
		where = append(where, "(e.title LIKE ? OR e.venue LIKE ? OR e.description LIKE ?)")
		args = append(args, like, like, like)
	}
	switch f.Scope {
	case "past":
		where = append(where, "COALESCE(e.ends_at, e.starts_at) < ?")
		args = append(args, s.now().UTC())
	case "all":
	default:
		if publicOnly || f.Scope == "upcoming" {
			where = append(where, "COALESCE(e.ends_at, e.starts_at) >= ?")
			args = append(args, s.now().UTC())
			order = "e.starts_at ASC"
		}
	}
	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	page := EventPage{Items: []Event{}, Limit: f.Limit, Offset: f.Offset}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alumni_events e`+clause, args...).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("count events: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, eventColumns+clause+" ORDER BY "+order+", e.id LIMIT ? OFFSET ?",
		append(args, f.Limit, f.Offset)...)
	if err != nil {
		return page, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return page, fmt.Errorf("scan event: %w", err)
		}
		page.Items = append(page.Items, e)
	}
	return page, rows.Err()
}
