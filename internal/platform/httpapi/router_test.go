package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type readinessFunc func(context.Context) error

func (f readinessFunc) PingContext(ctx context.Context) error {
	return f(ctx)
}

func TestHealthLiveDoesNotRequireDatabase(t *testing.T) {
	router := NewRouter(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	assertContentStatus(t, recorder.Body.Bytes(), "ok")
}

func TestHealthReadyReportsDatabaseFailureWithoutDetails(t *testing.T) {
	router := NewRouter(
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		readinessFunc(func(context.Context) error { return errors.New("private database detail") }),
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	assertContentStatus(t, recorder.Body.Bytes(), "not_ready")
	if strings.Contains(recorder.Body.String(), "private database detail") {
		t.Fatal("health response leaked the database error")
	}
}

func TestHealthReadyReportsReady(t *testing.T) {
	router := NewRouter(
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		readinessFunc(func(context.Context) error { return nil }),
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	assertContentStatus(t, recorder.Body.Bytes(), "ready")
}

func TestRequestLoggerOmitsQueryParameters(t *testing.T) {
	var output bytes.Buffer
	router := NewRouter(slog.New(slog.NewJSONHandler(&output, nil)), nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health/live?token=must-not-be-logged", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if strings.Contains(output.String(), "must-not-be-logged") {
		t.Fatal("request logger recorded a query parameter")
	}
	if !strings.Contains(output.String(), "/api/v1/health/live") {
		t.Fatal("request logger omitted the route")
	}
}

func TestRecoveryDoesNotExposePanicDetails(t *testing.T) {
	var output bytes.Buffer
	router := NewRouter(slog.New(slog.NewJSONHandler(&output, nil)), nil)
	router.Get("/api/v1/panic", func(http.ResponseWriter, *http.Request) {
		panic("private panic detail")
	})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/panic", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if strings.Contains(recorder.Body.String(), "private panic detail") || strings.Contains(output.String(), "private panic detail") {
		t.Fatal("panic details were exposed in the response or logs")
	}
}

func assertContentStatus(t *testing.T, body []byte, want string) {
	t.Helper()
	var result struct {
		Content struct {
			Status string `json:"status"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if result.Content.Status != want {
		t.Fatalf("content.status = %q, want %q", result.Content.Status, want)
	}
}
