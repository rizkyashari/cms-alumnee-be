package content

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	maxCommentRunes   = 1000
	commentsPerWindow = 10
	commentWindow     = time.Minute
)

var (
	ErrCommentNotFound  = errors.New("comment not found")
	ErrCommentForbidden = errors.New("only the author or a moderator may change this comment")
	ErrEventNotOpen     = errors.New("comments are only open on published events")
	ErrRateLimited      = errors.New("too many comments, try again shortly")
)

type Comment struct {
	ID         string  `json:"id"`
	Body       string  `json:"body"`
	AuthorName string  `json:"author_name"`
	AuthorRole string  `json:"author_role"` // verified | officer | admin | "" (label shown in the UI)
	CreatedAt  string  `json:"created_at"`
	EditedAt   *string `json:"edited_at"`
	CanChange  bool    `json:"can_change"` // viewer may edit or delete (author or moderator)
}

type CommentPage struct {
	Items  []Comment `json:"items"`
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

// commentLimiter is a per-account sliding window; like the login limiter it is
// in memory, so it bounds one process rather than the whole fleet.
type commentLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func (l *commentLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.attempts == nil {
		l.attempts = make(map[string][]time.Time)
	}
	recent := l.attempts[key][:0]
	for _, at := range l.attempts[key] {
		if now.Sub(at) < commentWindow {
			recent = append(recent, at)
		}
	}
	if len(recent) >= commentsPerWindow {
		l.attempts[key] = recent
		return false
	}
	l.attempts[key] = append(recent, now)
	return true
}

// cleanCommentBody trims the text and drops control characters other than
// newlines. Output is JSON and rendered as text by the client, never as HTML.
func cleanCommentBody(raw string) (string, error) {
	if !utf8.ValidString(raw) {
		return "", ValidationError{"comment must be valid UTF-8 text"}
	}
	cleaned := strings.Map(func(r rune) rune {
		if r == '\n' || !unicode.IsControl(r) {
			return r
		}
		if r == '\r' || r == '\t' {
			return ' '
		}
		return -1
	}, raw)
	cleaned = strings.TrimSpace(cleaned)
	if n := utf8.RuneCountInString(cleaned); n == 0 || n > maxCommentRunes {
		return "", ValidationError{"comment must be 1-1000 characters"}
	}
	return cleaned, nil
}

const commentColumns = `
	SELECT c.id, c.body, COALESCE(p.full_name, a.name, ''),
		CASE
			WHEN EXISTS (SELECT 1 FROM account_roles r WHERE r.account_id = c.author_id AND r.role_code = 'admin') THEN 'admin'
			WHEN EXISTS (SELECT 1 FROM account_roles r WHERE r.account_id = c.author_id AND r.role_code = 'officer') THEN 'officer'
			WHEN p.verification_status = 'verified' THEN 'verified'
			ELSE ''
		END,
		c.created_at, c.edited_at, c.author_id
	FROM event_comments c
	LEFT JOIN accounts a ON a.id = c.author_id
	LEFT JOIN alumni_profiles p ON p.account_id = c.author_id`

func scanComment(row rowScanner, viewer []byte, moderator bool) (Comment, error) {
	var (
		c         Comment
		id, owner []byte
		createdAt time.Time
		editedAt  sql.NullTime
	)
	if err := row.Scan(&id, &c.Body, &c.AuthorName, &c.AuthorRole, &createdAt, &editedAt, &owner); err != nil {
		return Comment{}, err
	}
	c.ID = formatID(id)
	c.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	if editedAt.Valid {
		v := editedAt.Time.UTC().Format(time.RFC3339)
		c.EditedAt = &v
	}
	c.CanChange = moderator || isOwner(owner, viewer)
	return c, nil
}

func (s *Service) ListComments(ctx context.Context, eventID, viewer []byte, moderator bool, limit, offset int) (CommentPage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	page := CommentPage{Items: []Comment{}, Limit: limit, Offset: offset}
	var status string
	err := s.db.QueryRowContext(ctx, `SELECT status FROM alumni_events WHERE id = ?`, eventID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && status != StatusPublished && !moderator) {
		return page, ErrEventNotFound
	}
	if err != nil {
		return page, fmt.Errorf("load event status: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM event_comments WHERE event_id = ?`, eventID).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("count comments: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, commentColumns+` WHERE c.event_id = ? ORDER BY c.created_at ASC, c.id LIMIT ? OFFSET ?`, eventID, limit, offset)
	if err != nil {
		return page, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		c, err := scanComment(rows, viewer, moderator)
		if err != nil {
			return page, fmt.Errorf("scan comment: %w", err)
		}
		page.Items = append(page.Items, c)
	}
	return page, rows.Err()
}

func (s *Service) loadComment(ctx context.Context, id, viewer []byte, moderator bool) (Comment, error) {
	c, err := scanComment(s.db.QueryRowContext(ctx, commentColumns+` WHERE c.id = ?`, id), viewer, moderator)
	if errors.Is(err, sql.ErrNoRows) {
		return Comment{}, ErrCommentNotFound
	}
	if err != nil {
		return Comment{}, fmt.Errorf("load comment: %w", err)
	}
	return c, nil
}

func (s *Service) AddComment(ctx context.Context, eventID, actor []byte, moderator bool, body string) (Comment, error) {
	cleaned, err := cleanCommentBody(body)
	if err != nil {
		return Comment{}, err
	}
	if !s.commentLimit.allow(string(actor), s.now()) {
		return Comment{}, ErrRateLimited
	}
	id, err := randomID()
	if err != nil {
		return Comment{}, fmt.Errorf("generate comment ID: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Comment{}, fmt.Errorf("begin add comment: %w", err)
	}
	defer tx.Rollback()
	_, status, err := lockEvent(ctx, tx, eventID)
	if err != nil {
		return Comment{}, err
	}
	if status != StatusPublished {
		return Comment{}, ErrEventNotOpen
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO event_comments (id, event_id, author_id, body) VALUES (?, ?, ?, ?)`,
		id, eventID, actor, cleaned); err != nil {
		return Comment{}, fmt.Errorf("insert comment: %w", err)
	}
	if err := audit(ctx, tx, actor, id, "event.comment.create", map[string]string{"event_id": formatID(eventID)}, "event_comment"); err != nil {
		return Comment{}, err
	}
	if err := tx.Commit(); err != nil {
		return Comment{}, fmt.Errorf("commit add comment: %w", err)
	}
	return s.loadComment(ctx, id, actor, moderator)
}

// changeComment locks the comment (which must belong to eventID) and checks
// that actor is its author or a moderator before running change.
func (s *Service) changeComment(ctx context.Context, eventID, commentID, actor []byte, moderator bool,
	change func(tx *sql.Tx, moderated bool) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin change comment: %w", err)
	}
	defer tx.Rollback()
	var owner []byte
	err = tx.QueryRowContext(ctx, `SELECT author_id FROM event_comments WHERE id = ? AND event_id = ? FOR UPDATE`, commentID, eventID).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCommentNotFound
	}
	if err != nil {
		return fmt.Errorf("lock comment: %w", err)
	}
	own := isOwner(owner, actor)
	if !own && !moderator {
		return ErrCommentForbidden
	}
	if err := change(tx, !own); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit change comment: %w", err)
	}
	return nil
}

func (s *Service) EditComment(ctx context.Context, eventID, commentID, actor []byte, moderator bool, body string) (Comment, error) {
	cleaned, err := cleanCommentBody(body)
	if err != nil {
		return Comment{}, err
	}
	err = s.changeComment(ctx, eventID, commentID, actor, moderator, func(tx *sql.Tx, moderated bool) error {
		if _, err := tx.ExecContext(ctx, `UPDATE event_comments SET body = ?, edited_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, cleaned, commentID); err != nil {
			return fmt.Errorf("update comment: %w", err)
		}
		return audit(ctx, tx, actor, commentID, "event.comment.update", map[string]string{"moderated": fmt.Sprint(moderated)}, "event_comment")
	})
	if err != nil {
		return Comment{}, err
	}
	return s.loadComment(ctx, commentID, actor, moderator)
}

func (s *Service) DeleteComment(ctx context.Context, eventID, commentID, actor []byte, moderator bool) error {
	return s.changeComment(ctx, eventID, commentID, actor, moderator, func(tx *sql.Tx, moderated bool) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM event_comments WHERE id = ?`, commentID); err != nil {
			return fmt.Errorf("delete comment: %w", err)
		}
		return audit(ctx, tx, actor, commentID, "event.comment.delete", map[string]string{"moderated": fmt.Sprint(moderated)}, "event_comment")
	})
}
