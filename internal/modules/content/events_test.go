package content

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rizkyashari/cms-alumnee-be/internal/modules/auth"
)

func validInput() EventInput {
	return EventInput{
		Title: "Temu kangen", Category: "Reuni", Venue: "Aula", Description: "Kumpul bareng",
		StartsAt: "2030-06-22T09:00:00+07:00",
	}
}

func TestValidateInput(t *testing.T) {
	end := "2030-06-22T08:00:00+07:00"
	badEnd := "nope"
	tests := []struct {
		name   string
		mutate func(*EventInput)
		reason string
	}{
		{"valid", func(*EventInput) {}, ""},
		{"short title", func(in *EventInput) { in.Title = "ab" }, "title"},
		{"unknown category", func(in *EventInput) { in.Category = "party" }, "category"},
		{"empty description", func(in *EventInput) { in.Description = "  " }, "description"},
		{"bad start", func(in *EventInput) { in.StartsAt = "besok" }, "starts_at"},
		{"end before start", func(in *EventInput) { in.EndsAt = &end }, "ends_at must be after"},
		{"bad end", func(in *EventInput) { in.EndsAt = &badEnd }, "ends_at must be an RFC"},
		{"long venue", func(in *EventInput) { in.Venue = strings.Repeat("x", 201) }, "venue"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := validInput()
			test.mutate(&in)
			n, err := validateInput(in)
			if test.reason == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if n.category != "reuni" || n.startsAt.Location().String() != "UTC" {
					t.Fatalf("input was not normalized: %+v", n)
				}
				return
			}
			if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("error = %v, want validation error containing %q", err, test.reason)
			}
		})
	}
}

func TestEscapeLike(t *testing.T) {
	if got := escapeLike(`50%_a\b`); got != `50\%\_a\\b` {
		t.Fatalf("escapeLike = %q", got)
	}
}

// Without a session every mutating or review route must refuse before any
// database access (the service has a nil *sql.DB).
func TestRoutesRequireAuthentication(t *testing.T) {
	authHandler := auth.NewHandler(auth.NewService(nil, false, "x"), slog.Default())
	router := chi.NewRouter()
	NewHandler(NewService(nil), authHandler, slog.Default()).RegisterRoutes(router)

	id := "11111111-1111-4111-8111-111111111111"
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/me/events"},
		{http.MethodPost, "/api/v1/events"},
		{http.MethodPut, "/api/v1/events/" + id},
		{http.MethodDelete, "/api/v1/events/" + id},
		{http.MethodPost, "/api/v1/events/" + id + "/submit"},
		{http.MethodGet, "/api/v1/admin/events"},
		{http.MethodPost, "/api/v1/admin/events/" + id + "/publish"},
		{http.MethodGet, "/api/v1/events/" + id + "/registration"},
		{http.MethodPost, "/api/v1/events/" + id + "/registration"},
		{http.MethodDelete, "/api/v1/events/" + id + "/registration"},
		{http.MethodGet, "/api/v1/me/registrations"},
		{http.MethodGet, "/api/v1/events/" + id + "/attendees"},
		{http.MethodPost, "/api/v1/events/" + id + "/attendees/" + id + "/attend"},
		{http.MethodPost, "/api/v1/events/" + id + "/comments"},
		{http.MethodPut, "/api/v1/events/" + id + "/comments/" + id},
		{http.MethodDelete, "/api/v1/events/" + id + "/comments/" + id},
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(route.method, route.path, nil))
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", route.method, route.path, recorder.Code)
		}
	}
}

func TestInvalidEventIDIsNotFound(t *testing.T) {
	authHandler := auth.NewHandler(auth.NewService(nil, false, "x"), slog.Default())
	router := chi.NewRouter()
	NewHandler(NewService(nil), authHandler, slog.Default()).RegisterRoutes(router)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/events/not-an-id", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func TestCleanCommentBody(t *testing.T) {
	got, err := cleanCommentBody("  halo\x00 <b>dunia</b>\r\nbaris dua\x07 ")
	if err != nil || !strings.Contains(got, "<b>dunia</b>") || !strings.Contains(got, "\nbaris dua") {
		t.Fatalf("unexpected result %q, %v", got, err)
	}
	if strings.ContainsAny(got, "\x00\x07\r") {
		t.Fatalf("control characters survived: %q", got)
	}
	for _, bad := range []string{"", "   \n ", strings.Repeat("a", 1001), "\xff\xfe"} {
		if _, err := cleanCommentBody(bad); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("cleanCommentBody(%q) error = %v, want validation error", bad, err)
		}
	}
	if _, err := cleanCommentBody(strings.Repeat("é", 1000)); err != nil {
		t.Errorf("1000 multi-byte characters rejected: %v", err)
	}
}

func TestCommentLimiter(t *testing.T) {
	var limiter commentLimiter
	now := time.Now()
	for i := 0; i < commentsPerWindow; i++ {
		if !limiter.allow("a", now) {
			t.Fatalf("comment %d refused", i)
		}
	}
	if limiter.allow("a", now) {
		t.Fatal("limit not enforced")
	}
	if !limiter.allow("b", now) {
		t.Fatal("limit leaked across accounts")
	}
	if !limiter.allow("a", now.Add(commentWindow+time.Second)) {
		t.Fatal("window did not expire")
	}
}
