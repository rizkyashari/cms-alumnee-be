package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

type principalContextKey struct{}

type Handler struct {
	service         *Service
	limiter         *loginLimiter
	otpIPLimiter    *loginLimiter
	otpEmailLimiter *loginLimiter
	logger          *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service, limiter: newLoginLimiter(5, time.Minute),
		otpIPLimiter:    newLoginLimiter(10, time.Hour),
		otpEmailLimiter: newLoginLimiter(3, 15*time.Minute),
		logger:          logger,
	}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	registerEmailOTPRoutes(h, router)
	h.registerVerificationRoutes(router)
	h.registerRegionRoutes(router)
	router.With(h.RequireCapability(CapabilityAdminManage)).Get("/api/v1/admin/roles", h.managedRoles)
	router.With(h.RequireCapability(CapabilityAdminManage)).Post("/api/v1/admin/roles/{operation}", h.changeManagedRole)
	router.Post("/api/v1/auth/demo-login", h.demoLogin)
	router.Post("/api/v1/auth/password-login", h.passwordLogin)
	h.registerOAuthRoutes(router)
	router.With(h.Authenticate).Get("/api/v1/auth/me", h.me)
	router.With(h.RequireAnyCapability(CapabilityDirectoryLite, CapabilityDirectoryFull)).Get("/api/v1/alumni", h.directory)
	router.Post("/api/v1/auth/logout", h.logout)
}

func (h *Handler) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := h.service.Authenticate(r)
		if err != nil {
			if !errors.Is(err, ErrUnauthenticated) {
				h.reportFailure("load_session", err)
				writeError(w, http.StatusServiceUnavailable, "authentication service unavailable")
				return
			}
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalContextKey{}, principal)))
	})
}

func (h *Handler) RequireCapability(capability Capability) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if capability == CapabilityPublicContent {
			return RequireCapability(capability)(next)
		}
		return h.Authenticate(RequireCapability(capability)(next))
	}
}

func (h *Handler) RequireAnyCapability(capabilities ...Capability) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for _, capability := range capabilities {
			if capability == CapabilityPublicContent {
				return RequireAnyCapability(capabilities...)(next)
			}
		}
		return h.Authenticate(RequireAnyCapability(capabilities...)(next))
	}
}

func RequireCapability(capability Capability) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := r.Context().Value(principalContextKey{}).(Principal)
			if !ok {
				if capability == CapabilityPublicContent {
					next.ServeHTTP(w, r)
					return
				}
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if roleHasCapability(principal.Role, capability) {
				next.ServeHTTP(w, r)
				return
			}
			writeError(w, http.StatusForbidden, "forbidden")
		})
	}
}

func RequireAnyCapability(capabilities ...Capability) func(http.Handler) http.Handler {
	allowed := make(map[Capability]struct{}, len(capabilities))
	for _, capability := range capabilities {
		allowed[capability] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFromContext(r.Context())
			if !ok {
				if _, publicAllowed := allowed[CapabilityPublicContent]; publicAllowed {
					next.ServeHTTP(w, r)
					return
				}
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			for capability := range allowed {
				if roleHasCapability(principal.Role, capability) {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeError(w, http.StatusForbidden, "forbidden")
		})
	}
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

func roleHasCapability(role Role, capability Capability) bool {
	for _, granted := range CapabilitiesFor(role) {
		if granted == capability {
			return true
		}
	}
	return false
}

func (h *Handler) directory(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if principal.Role == RoleNonVerified {
		rows, err := h.service.db.QueryContext(r.Context(), `
			SELECT ap.full_name, ap.batch_year
			FROM alumni_profiles ap
			JOIN accounts a ON a.id = ap.account_id
			WHERE a.status = 'active'
			ORDER BY ap.batch_year DESC, ap.full_name
			LIMIT 100`)
		if err != nil {
			h.reportFailure("read_limited_directory", err)
			writeError(w, http.StatusInternalServerError, "could not load directory")
			return
		}
		defer rows.Close()
		items := make([]map[string]any, 0)
		for rows.Next() {
			var name string
			var batchYear int
			if err := rows.Scan(&name, &batchYear); err != nil {
				h.reportFailure("read_limited_directory", err)
				writeError(w, http.StatusInternalServerError, "could not load directory")
				return
			}
			items = append(items, map[string]any{"name": name, "batch_year": batchYear})
		}
		if err := rows.Err(); err != nil {
			h.reportFailure("read_limited_directory", err)
			writeError(w, http.StatusInternalServerError, "could not load directory")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"message": "ok", "content": map[string]any{"items": items}})
		return
	}

	rows, err := h.service.db.QueryContext(r.Context(), `
		SELECT ap.full_name, ap.batch_year, ap.city, COALESCE(ap.occupation, ''),
			COALESCE(ap.bio, ''), COALESCE(c.value, '')
		FROM alumni_profiles ap
		JOIN accounts a ON a.id = ap.account_id
		LEFT JOIN account_contacts c
			ON c.account_id = a.id AND c.contact_type = 'phone'
			AND c.is_primary = TRUE AND c.show_in_directory = TRUE AND c.verified_at IS NOT NULL
		WHERE a.status = 'active'
		ORDER BY ap.batch_year DESC, ap.full_name
		LIMIT 100`)
	if err != nil {
		h.reportFailure("read_full_directory", err)
		writeError(w, http.StatusInternalServerError, "could not load directory")
		return
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var name, city, occupation, bio, phone string
		var batchYear int
		if err := rows.Scan(&name, &batchYear, &city, &occupation, &bio, &phone); err != nil {
			h.reportFailure("read_full_directory", err)
			writeError(w, http.StatusInternalServerError, "could not load directory")
			return
		}
		item := map[string]any{
			"name": name, "batch_year": batchYear, "city": city,
			"occupation": occupation, "bio": bio,
		}
		if phone != "" {
			item["phone"] = phone
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		h.reportFailure("read_full_directory", err)
		writeError(w, http.StatusInternalServerError, "could not load directory")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "ok", "content": map[string]any{"items": items}})
}

func (h *Handler) demoLogin(w http.ResponseWriter, r *http.Request) {
	if h.service.demoAuthEnabled && !isLoopbackRequest(r) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if !h.limiter.Allow(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "too many login attempts")
		return
	}
	var input LoginInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	principal, token, expiresAt, err := h.service.DemoLogin(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrDemoAuthDisabled) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		h.reportFailure("demo_login", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}
	h.service.SetSessionCookie(w, token, expiresAt)
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "login successful",
		"content": map[string]any{"user": principal, "expires_at": expiresAt},
	})
}

func (h *Handler) passwordLogin(w http.ResponseWriter, r *http.Request) {
	if !h.service.passwordLoginEnabled {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if !h.service.sameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	if !h.limiter.Allow(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "too many login attempts")
		return
	}
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	principal, token, expiresAt, err := h.service.PasswordLogin(r.Context(), input.Username, input.Password)
	if err != nil {
		if errors.Is(err, ErrPasswordAuthDisabled) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		h.reportFailure("password_login", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}
	h.service.SetSessionCookie(w, token, expiresAt)
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "login successful",
		"content": map[string]any{"user": principal, "expires_at": expiresAt},
	})
}

func isLoopbackRequest(r *http.Request) bool {
	if !isLoopbackHost(r.Host) {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsedOrigin, err := url.Parse(origin)
	if err != nil || (parsedOrigin.Scheme != "http" && parsedOrigin.Scheme != "https") {
		return false
	}
	return isLoopbackHost(parsedOrigin.Host)
}

func isLoopbackHost(address string) bool {
	host := address
	if parsedHost, _, err := net.SplitHostPort(address); err == nil {
		host = parsedHost
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	completion, err := h.service.profileCompletion(r.Context(), principalIDBytes(principal.AccountID))
	if err != nil {
		h.reportFailure("profile_completion", err)
		writeError(w, http.StatusInternalServerError, "could not load profile completion")
		return
	}
	principal.ProfileCompletion = completion
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{"message": "ok", "content": map[string]any{"user": principal}})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if !h.service.sameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		if err := h.service.RevokeSession(r.Context(), cookie.Value); err != nil {
			h.reportFailure("logout", err)
			writeError(w, http.StatusInternalServerError, "logout failed")
			return
		}
	}
	h.service.ClearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{"message": "logged out", "content": nil})
}

func (h *Handler) reportFailure(operation string, err error) {
	if h.logger == nil {
		return
	}
	h.logger.Error("authentication operation failed",
		slog.String("operation", operation),
		slog.String("error_type", fmt.Sprintf("%T", err)),
	)
}

func (s *Service) sameOriginRequest(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return r.Header.Get("Sec-Fetch-Site") == "same-origin" || r.Header.Get("Sec-Fetch-Site") == "none"
	}
	originURL, err := url.Parse(origin)
	if err != nil || (originURL.Scheme != "http" && originURL.Scheme != "https") || originURL.Host == "" ||
		originURL.User != nil || originURL.Path != "" || originURL.RawQuery != "" || originURL.Fragment != "" {
		return false
	}
	if s.demoAuthEnabled {
		return isLoopbackHost(r.Host) && isLoopbackHost(originURL.Host)
	}
	if s.integrations.PublicBaseURL == "" {
		return false
	}
	publicURL, err := url.Parse(s.integrations.PublicBaseURL)
	return err == nil && sameURLOrigin(originURL, publicURL)
}

func sameURLOrigin(left, right *url.URL) bool {
	return strings.EqualFold(left.Scheme, right.Scheme) &&
		strings.EqualFold(left.Hostname(), right.Hostname()) &&
		effectivePort(left) == effectivePort(right)
}

func effectivePort(value *url.URL) string {
	if port := value.Port(); port != "" {
		return port
	}
	if strings.EqualFold(value.Scheme, "https") {
		return "443"
	}
	return "80"
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request must contain one JSON value")
	}
	return nil
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "unknown"
	}
	return host
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"message": message, "content": nil})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

type loginLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
	now      func() time.Time
}

func newLoginLimiter(limit int, window time.Duration) *loginLimiter {
	return &loginLimiter{attempts: make(map[string][]time.Time), limit: limit, window: window, now: time.Now}
}

func (l *loginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	recent := l.attempts[key][:0]
	for _, attempt := range l.attempts[key] {
		if now.Sub(attempt) < l.window {
			recent = append(recent, attempt)
		}
	}
	if _, exists := l.attempts[key]; !exists && len(l.attempts) >= 10000 {
		for address, attempts := range l.attempts {
			if len(attempts) == 0 || now.Sub(attempts[len(attempts)-1]) >= l.window {
				delete(l.attempts, address)
			}
		}
		if len(l.attempts) >= 10000 {
			return false
		}
	}
	if len(recent) >= l.limit {
		l.attempts[key] = recent
		return false
	}
	l.attempts[key] = append(recent, now)
	return true
}
