package content

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rizkyashari/cms-alumnee-be/internal/platform/migrations"
)

// Needs a disposable MySQL 8 database: set TEST_MYSQL_DSN (with parseTime=true).
// It deletes all events and accounts, so never point it at a shared database,
// and run with `go test -p 1 ./...` because packages share that one database.
func openDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := migrations.Up(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	for _, statement := range []string{`DELETE FROM audit_events`, `DELETE FROM alumni_events`, `DELETE FROM accounts`} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func seedAccount(t *testing.T, db *sql.DB, name string) []byte {
	t.Helper()
	id, _ := randomID()
	if _, err := db.Exec(`INSERT INTO accounts (id, name) VALUES (?, ?)`, id, name); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestEventReviewWorkflow(t *testing.T) {
	db := openDB(t)
	s := NewService(db)
	ctx := context.Background()
	author, other, admin := seedAccount(t, db, "Penulis"), seedAccount(t, db, "Lain"), seedAccount(t, db, "Admin")
	in := validInput()
	in.StartsAt = time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)

	event, err := s.CreateDraft(ctx, author, in)
	if err != nil || event.Status != StatusDraft {
		t.Fatalf("create draft: %v %+v", err, event)
	}
	id := parseID(event.ID)

	// Drafts are never public and cannot be touched by another member.
	if _, err := s.GetPublished(ctx, id); !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("draft visible publicly: %v", err)
	}
	if _, err := s.UpdateOwn(ctx, other, id, in); !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("other member edited draft: %v", err)
	}
	if _, err := s.Decide(ctx, admin, id, "publish", ""); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("published a draft directly: %v", err)
	}

	if err := s.Submit(ctx, author, id); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if err := s.Submit(ctx, author, id); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("double submit: %v", err)
	}
	if _, err := s.UpdateOwn(ctx, author, id, in); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("edited while in review: %v", err)
	}

	if _, err := s.Decide(ctx, admin, id, "reject", ""); !errors.Is(err, ErrNoteRequired) {
		t.Fatalf("reject without note: %v", err)
	}
	rejected, err := s.Decide(ctx, admin, id, "reject", "Lengkapi lokasi")
	if err != nil || rejected.Status != StatusRejected || rejected.ReviewNote != "Lengkapi lokasi" {
		t.Fatalf("reject: %v %+v", err, rejected)
	}

	// Editing a rejected event returns it to draft for resubmission.
	if edited, err := s.UpdateOwn(ctx, author, id, in); err != nil || edited.Status != StatusDraft || edited.ReviewNote != "" {
		t.Fatalf("edit rejected: %v %+v", err, edited)
	}
	if err := s.Submit(ctx, author, id); err != nil {
		t.Fatalf("resubmit: %v", err)
	}
	published, err := s.Decide(ctx, admin, id, "publish", "")
	if err != nil || published.Status != StatusPublished || published.PublishedAt == nil {
		t.Fatalf("publish: %v %+v", err, published)
	}

	page, err := s.List(ctx, ListFilter{}, true, nil)
	if err != nil || page.Total != 1 || page.Items[0].ID != event.ID {
		t.Fatalf("public list: %v %+v", err, page)
	}
	if page, _ = s.List(ctx, ListFilter{Category: "karier"}, true, nil); page.Total != 0 {
		t.Fatalf("category filter returned %d", page.Total)
	}
	if page, _ = s.List(ctx, ListFilter{Query: "100%"}, true, nil); page.Total != 0 {
		t.Fatalf("wildcard search matched %d", page.Total)
	}
	if err := s.DeleteOwn(ctx, author, id); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("deleted a published event: %v", err)
	}

	// Comments: author or moderator may change them, other members may not.
	if _, err := s.AddComment(ctx, id, other, false, "Siap hadir"); err != nil {
		t.Fatalf("comment on published event: %v", err)
	}
	comments, err := s.ListComments(ctx, id, other, false, 0, 0)
	if err != nil || comments.Total != 1 || !comments.Items[0].CanChange || comments.Items[0].AuthorName != "Lain" {
		t.Fatalf("list comments: %v %+v", err, comments)
	}
	commentID := parseID(comments.Items[0].ID)
	if as, _ := s.ListComments(ctx, id, author, false, 0, 0); as.Items[0].CanChange {
		t.Fatal("non-author sees CanChange")
	}
	if _, err := s.EditComment(ctx, id, commentID, author, false, "bukan saya"); !errors.Is(err, ErrCommentForbidden) {
		t.Fatalf("other member edited comment: %v", err)
	}
	if edited, err := s.EditComment(ctx, id, commentID, admin, true, "Dimoderasi"); err != nil || edited.EditedAt == nil {
		t.Fatalf("moderator edit: %v %+v", err, edited)
	}
	if err := s.DeleteComment(ctx, id, commentID, other, false); err != nil {
		t.Fatalf("author delete: %v", err)
	}
	if err := s.DeleteComment(ctx, id, commentID, other, false); !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("double delete: %v", err)
	}
	if _, err := s.AddComment(ctx, parseID(event.ID), other, false, "   "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("blank comment: %v", err)
	}

	if _, err := s.AddComment(ctx, id, other, false, "setelah arsip"); err != nil {
		t.Fatalf("comment before archive: %v", err)
	}
	if _, err := s.Decide(ctx, admin, id, "archive", ""); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if _, err := s.GetPublished(ctx, id); !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("archived event still public: %v", err)
	}
	if _, err := s.AddComment(ctx, id, other, false, "terlambat"); !errors.Is(err, ErrEventNotOpen) {
		t.Fatalf("comment on archived event: %v", err)
	}

	var audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_events WHERE resource_type = 'event' AND resource_id = ?`, id).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	// event: create, submit, reject, update, submit, publish, archive; refused attempts write nothing.
	if audits != 7 {
		t.Fatalf("audit rows = %d, want 7", audits)
	}
}

func TestSubmitRejectsPastStart(t *testing.T) {
	db := openDB(t)
	s := NewService(db)
	author := seedAccount(t, db, "Penulis")
	in := validInput()
	in.StartsAt = time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	event, err := s.CreateDraft(context.Background(), author, in)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Submit(context.Background(), author, parseID(event.ID)); !errors.Is(err, ErrStartsInPast) {
		t.Fatalf("submit past event: %v", err)
	}
}

func TestCommentAuthorRoleLabels(t *testing.T) {
	db := openDB(t)
	s := NewService(db)
	ctx := context.Background()
	author, verified, officer, admin, plain := seedAccount(t, db, "Penulis"), seedAccount(t, db, "Verified"),
		seedAccount(t, db, "Pengurus"), seedAccount(t, db, "Admin"), seedAccount(t, db, "Biasa")
	for account, role := range map[*[]byte]string{&verified: "member", &officer: "officer", &admin: "admin", &plain: "member"} {
		if _, err := db.Exec(`INSERT INTO account_roles (account_id, role_code) VALUES (?, ?)`, *account, role); err != nil {
			t.Fatal(err)
		}
	}
	// Verified/plain members have profiles; an admin who is also a verified alumnus must still read "admin".
	for account, status := range map[*[]byte]string{&verified: "verified", &admin: "verified", &plain: "non_verified"} {
		profileID, _ := randomID()
		if _, err := db.Exec(`INSERT INTO alumni_profiles (id, account_id, full_name, batch_year, city, verification_status)
			VALUES (?, ?, 'Nama Profil', 2010, 'Yogyakarta', ?)`, profileID, *account, status); err != nil {
			t.Fatal(err)
		}
	}
	in := validInput()
	in.StartsAt = time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	event, err := s.CreateDraft(ctx, author, in)
	if err != nil {
		t.Fatal(err)
	}
	id := parseID(event.ID)
	if _, err := db.Exec(`UPDATE alumni_events SET status = 'published', published_at = NOW(6) WHERE id = ?`, id); err != nil {
		t.Fatal(err)
	}
	for _, commenter := range [][]byte{verified, officer, admin, plain} {
		if _, err := s.AddComment(ctx, id, commenter, false, "halo"); err != nil {
			t.Fatal(err)
		}
	}
	page, err := s.ListComments(ctx, id, author, false, 0, 0)
	if err != nil || page.Total != 4 {
		t.Fatalf("list: %v %+v", err, page)
	}
	roles := map[string]int{}
	for _, c := range page.Items {
		roles[c.AuthorRole]++
	}
	if roles["verified"] != 1 || roles["officer"] != 1 || roles["admin"] != 1 || roles[""] != 1 {
		t.Fatalf("role counts = %v, want one each of verified/officer/admin/none", roles)
	}
}

func seedProfile(t *testing.T, db *sql.DB, account []byte, name string) []byte {
	t.Helper()
	id, _ := randomID()
	if _, err := db.Exec(`INSERT INTO alumni_profiles (id, account_id, full_name, batch_year, city) VALUES (?, ?, ?, 2010, 'Yogyakarta')`, id, account, name); err != nil {
		t.Fatal(err)
	}
	return id
}

func publishedEvent(t *testing.T, db *sql.DB, s *Service, author []byte, starts, ends time.Time) []byte {
	t.Helper()
	in := validInput()
	in.StartsAt = starts.UTC().Format(time.RFC3339)
	e := ends.UTC().Format(time.RFC3339)
	in.EndsAt = &e
	event, err := s.CreateDraft(context.Background(), author, in)
	if err != nil {
		t.Fatal(err)
	}
	id := parseID(event.ID)
	if _, err := db.Exec(`UPDATE alumni_events SET status = 'published', published_at = NOW(6) WHERE id = ?`, id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestEventRegistrations(t *testing.T) {
	db := openDB(t)
	s := NewService(db)
	ctx := context.Background()
	author, member, noProfile, stranger := seedAccount(t, db, "Penulis"), seedAccount(t, db, "Peserta"), seedAccount(t, db, "Tanpa Profil"), seedAccount(t, db, "Orang Lain")
	memberProfile := seedProfile(t, db, member, "Peserta")
	seedProfile(t, db, stranger, "Orang Lain")
	now := time.Now()
	future := publishedEvent(t, db, s, author, now.Add(24*time.Hour), now.Add(26*time.Hour))
	finished := publishedEvent(t, db, s, author, now.Add(-4*time.Hour), now.Add(-2*time.Hour))
	started := publishedEvent(t, db, s, author, now.Add(-time.Hour), now.Add(time.Hour))

	if status, err := s.RegistrationStatus(ctx, future, member); err != nil || status != RegistrationNone {
		t.Fatalf("initial status: %v %q", err, status)
	}
	if _, err := s.Register(ctx, future, noProfile); !errors.Is(err, ErrProfileRequired) {
		t.Fatalf("register without profile: %v", err)
	}
	if _, err := s.Register(ctx, finished, member); !errors.Is(err, ErrEventFinished) {
		t.Fatalf("register for finished event: %v", err)
	}
	draft, _ := s.CreateDraft(ctx, author, validInput())
	if _, err := s.Register(ctx, parseID(draft.ID), member); !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("register for unpublished event: %v", err)
	}
	for i := 0; i < 2; i++ { // idempotent
		if status, err := s.Register(ctx, future, member); err != nil || status != RegistrationRegistered {
			t.Fatalf("register #%d: %v %q", i, err, status)
		}
	}
	if e, _ := s.GetPublished(ctx, future); e.AttendeeCount != 1 {
		t.Fatalf("attendee_count = %d, want 1", e.AttendeeCount)
	}
	if err := s.Unregister(ctx, future, member); err != nil {
		t.Fatalf("unregister: %v", err)
	}
	if e, _ := s.GetPublished(ctx, future); e.AttendeeCount != 0 {
		t.Fatalf("attendee_count after cancel = %d, want 0", e.AttendeeCount)
	}
	if err := s.Unregister(ctx, future, stranger); !errors.Is(err, ErrRegistrationNotFound) {
		t.Fatalf("unregister never registered: %v", err)
	}
	if status, err := s.Register(ctx, future, member); err != nil || status != RegistrationRegistered {
		t.Fatalf("re-register: %v %q", err, status)
	}

	mine, err := s.ListMine(ctx, member)
	if err != nil || len(mine) != 1 || mine[0].Event.ID != formatID(future) {
		t.Fatalf("list mine: %v %+v", err, mine)
	}

	// Attendance: only the author or a moderator, and only after the event starts.
	if _, err := s.Register(ctx, started, member); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ListAttendees(ctx, started, stranger, false); !errors.Is(err, ErrNotOrganizer) {
		t.Fatalf("stranger listed attendees: %v", err)
	}
	attendees, err := s.ListAttendees(ctx, started, author, false)
	if err != nil || len(attendees) != 1 || attendees[0].Name != "Peserta" {
		t.Fatalf("list attendees: %v %+v", err, attendees)
	}
	if err := s.MarkAttended(ctx, future, memberProfile, author, false); !errors.Is(err, ErrEventNotStarted) {
		t.Fatalf("attend before start: %v", err)
	}
	if err := s.MarkAttended(ctx, started, memberProfile, stranger, false); !errors.Is(err, ErrNotOrganizer) {
		t.Fatalf("stranger marked attendance: %v", err)
	}
	if err := s.MarkAttended(ctx, started, memberProfile, author, false); err != nil {
		t.Fatalf("mark attended: %v", err)
	}
	if err := s.Unregister(ctx, started, member); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("cancelled after attending: %v", err)
	}
	if status, _ := s.RegistrationStatus(ctx, started, member); status != RegistrationAttended {
		t.Fatalf("status = %q, want attended", status)
	}
}
