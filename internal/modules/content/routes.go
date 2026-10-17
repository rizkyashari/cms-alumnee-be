package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rizkyashari/cms-alumnee-be/internal/modules/auth"
)

const maxBodyBytes = 16 << 10

var reviewStatuses = map[string]struct{}{
	StatusDraft: {}, StatusPendingReview: {}, StatusPublished: {}, StatusRejected: {}, StatusArchived: {},
}

type Handler struct {
	service *Service
	auth    *auth.Handler
	logger  *slog.Logger
}

func NewHandler(service *Service, authHandler *auth.Handler, logger *slog.Logger) *Handler {
	return &Handler{service: service, auth: authHandler, logger: logger}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	submit := h.auth.RequireCapability(auth.CapabilityEventSubmit)
	review := h.auth.RequireCapability(auth.CapabilityContentReview)
	comment := h.auth.RequireCapability(auth.CapabilityComment)

	router.Get("/api/v1/events", h.listPublic)
	router.Get("/api/v1/events/{eventID}", h.getPublic)

	router.With(submit).Get("/api/v1/me/events", h.listOwn)
	router.With(submit).Post("/api/v1/events", h.create)
	router.With(submit).Put("/api/v1/events/{eventID}", h.update)
	router.With(submit).Delete("/api/v1/events/{eventID}", h.remove)
	router.With(submit).Post("/api/v1/events/{eventID}/submit", h.submit)

	router.With(h.auth.OptionalAuthenticate).Get("/api/v1/events/{eventID}/comments", h.listComments)
	router.With(comment).Post("/api/v1/events/{eventID}/comments", h.addComment)
	router.With(h.auth.Authenticate).Put("/api/v1/events/{eventID}/comments/{commentID}", h.editComment)
	router.With(h.auth.Authenticate).Delete("/api/v1/events/{eventID}/comments/{commentID}", h.deleteComment)

	register := h.auth.RequireCapability(auth.CapabilityEventRegister)
	router.With(register).Get("/api/v1/events/{eventID}/registration", h.registrationStatus)
	router.With(register).Post("/api/v1/events/{eventID}/registration", h.register)
	router.With(register).Delete("/api/v1/events/{eventID}/registration", h.unregister)
	router.With(register).Get("/api/v1/me/registrations", h.myRegistrations)
	router.With(submit).Get("/api/v1/events/{eventID}/attendees", h.attendees)
	router.With(submit).Post("/api/v1/events/{eventID}/attendees/{profileID}/attend", h.markAttended)

	router.With(review).Get("/api/v1/admin/events", h.listAdmin)
	router.With(review).Post("/api/v1/admin/events/{eventID}/{decision}", h.decide)
}

func (h *Handler) fail(operation string, err error) {
	if h.logger != nil {
		h.logger.Error("content operation failed", slog.String("operation", operation), slog.String("error_type", fmt.Sprintf("%T", err)))
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"message": message, "content": nil})
}

func writeContent(w http.ResponseWriter, status int, message string, content any) {
	writeJSON(w, status, map[string]any{"message": message, "content": content})
}

func decodeBody(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
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

func filterFrom(r *http.Request) ListFilter {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	return ListFilter{
		Query: q.Get("q"), Category: q.Get("category"), Scope: q.Get("scope"),
		Limit: limit, Offset: offset,
	}
}

func (h *Handler) listPublic(w http.ResponseWriter, r *http.Request) {
	page, err := h.service.List(r.Context(), filterFrom(r), true, nil)
	if err != nil {
		h.fail("list_events", err)
		writeError(w, http.StatusInternalServerError, "could not load events")
		return
	}
	writeContent(w, http.StatusOK, "ok", page)
}

func (h *Handler) getPublic(w http.ResponseWriter, r *http.Request) {
	id := parseID(chi.URLParam(r, "eventID"))
	if id == nil {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	event, err := h.service.GetPublished(r.Context(), id)
	if errors.Is(err, ErrEventNotFound) {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		h.fail("get_event", err)
		writeError(w, http.StatusInternalServerError, "could not load event")
		return
	}
	writeContent(w, http.StatusOK, "ok", event)
}

func (h *Handler) listOwn(w http.ResponseWriter, r *http.Request) {
	principal, _ := auth.PrincipalFromContext(r.Context())
	filter := filterFrom(r)
	filter.Scope = "all"
	if status := r.URL.Query().Get("status"); status != "" {
		if _, ok := reviewStatuses[status]; !ok {
			writeError(w, http.StatusBadRequest, "unknown status")
			return
		}
		filter.Status = status
	}
	page, err := h.service.List(r.Context(), filter, false, parseID(principal.AccountID))
	if err != nil {
		h.fail("list_own_events", err)
		writeError(w, http.StatusInternalServerError, "could not load events")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeContent(w, http.StatusOK, "ok", page)
}

func (h *Handler) listAdmin(w http.ResponseWriter, r *http.Request) {
	filter := filterFrom(r)
	filter.Scope = "all"
	filter.Status = StatusPendingReview
	if status := r.URL.Query().Get("status"); status != "" {
		if _, ok := reviewStatuses[status]; !ok {
			writeError(w, http.StatusBadRequest, "unknown status")
			return
		}
		filter.Status = status
	}
	page, err := h.service.List(r.Context(), filter, false, nil)
	if err != nil {
		h.fail("list_admin_events", err)
		writeError(w, http.StatusInternalServerError, "could not load events")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeContent(w, http.StatusOK, "ok", page)
}

// mutationContext rejects cross-origin requests and resolves the caller and
// target event; ok is false once a response has been written.
func (h *Handler) mutationContext(w http.ResponseWriter, r *http.Request, withEvent bool) (actor, id []byte, ok bool) {
	if !h.auth.SameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return nil, nil, false
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	actor = parseID(principal.AccountID)
	if withEvent {
		if id = parseID(chi.URLParam(r, "eventID")); id == nil {
			writeError(w, http.StatusNotFound, "event not found")
			return nil, nil, false
		}
	}
	return actor, id, true
}

// writeServiceError maps domain errors to HTTP responses; it returns false for
// unexpected errors after logging them and sending a 500.
func (h *Handler) writeServiceError(w http.ResponseWriter, operation string, err error) {
	var validation ValidationError
	switch {
	case errors.As(err, &validation):
		writeError(w, http.StatusUnprocessableEntity, validation.Reason)
	case errors.Is(err, ErrEventNotFound):
		writeError(w, http.StatusNotFound, "event not found")
	case errors.Is(err, ErrInvalidState):
		writeError(w, http.StatusConflict, "event is not in a state that allows this action")
	case errors.Is(err, ErrStartsInPast):
		writeError(w, http.StatusUnprocessableEntity, "event must start in the future")
	case errors.Is(err, ErrNoteRequired):
		writeError(w, http.StatusUnprocessableEntity, "a note is required when rejecting an event")
	case errors.Is(err, ErrProfileRequired):
		writeError(w, http.StatusConflict, "complete your alumni profile before registering")
	case errors.Is(err, ErrEventFinished):
		writeError(w, http.StatusConflict, "event has already finished")
	case errors.Is(err, ErrEventNotStarted):
		writeError(w, http.StatusConflict, "event has not started yet")
	case errors.Is(err, ErrRegistrationNotFound):
		writeError(w, http.StatusNotFound, "registration not found")
	case errors.Is(err, ErrNotOrganizer):
		writeError(w, http.StatusForbidden, "only the event author or an Admin may manage attendance")
	case errors.Is(err, ErrCommentNotFound):
		writeError(w, http.StatusNotFound, "comment not found")
	case errors.Is(err, ErrCommentForbidden):
		writeError(w, http.StatusForbidden, "only the author or a moderator may change this comment")
	case errors.Is(err, ErrEventNotOpen):
		writeError(w, http.StatusConflict, "comments are only open on published events")
	case errors.Is(err, ErrRateLimited):
		writeError(w, http.StatusTooManyRequests, "too many comments, try again shortly")
	case errors.Is(err, ErrUnknownDecison):
		writeError(w, http.StatusNotFound, "unknown decision")
	default:
		h.fail(operation, err)
		writeError(w, http.StatusInternalServerError, "could not complete the request")
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	actor, _, ok := h.mutationContext(w, r, false)
	if !ok {
		return
	}
	var input EventInput
	if err := decodeBody(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	event, err := h.service.CreateDraft(r.Context(), actor, input)
	if err != nil {
		h.writeServiceError(w, "create_event", err)
		return
	}
	writeContent(w, http.StatusCreated, "draft saved", event)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.mutationContext(w, r, true)
	if !ok {
		return
	}
	var input EventInput
	if err := decodeBody(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	event, err := h.service.UpdateOwn(r.Context(), actor, id, input)
	if err != nil {
		h.writeServiceError(w, "update_event", err)
		return
	}
	writeContent(w, http.StatusOK, "draft saved", event)
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.mutationContext(w, r, true)
	if !ok {
		return
	}
	if err := h.service.Submit(r.Context(), actor, id); err != nil {
		h.writeServiceError(w, "submit_event", err)
		return
	}
	writeContent(w, http.StatusAccepted, "event submitted for review", nil)
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.mutationContext(w, r, true)
	if !ok {
		return
	}
	if err := h.service.DeleteOwn(r.Context(), actor, id); err != nil {
		h.writeServiceError(w, "delete_event", err)
		return
	}
	writeContent(w, http.StatusOK, "event deleted", nil)
}

func (h *Handler) decide(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := h.mutationContext(w, r, true)
	if !ok {
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	if r.ContentLength != 0 {
		if err := decodeBody(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}
	event, err := h.service.Decide(r.Context(), actor, id, chi.URLParam(r, "decision"), body.Note)
	if err != nil {
		h.writeServiceError(w, "decide_event", err)
		return
	}
	writeContent(w, http.StatusOK, "event "+event.Status, event)
}

// commentContext resolves the caller, moderator status and target IDs for the
// comment routes; ok is false once a response has been written.
func (h *Handler) commentContext(w http.ResponseWriter, r *http.Request, mutating, withComment bool) (actor, eventID, commentID []byte, moderator, ok bool) {
	if mutating && !h.auth.SameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	actor = parseID(principal.AccountID)
	for _, capability := range principal.Capabilities {
		if capability == auth.CapabilityContentReview {
			moderator = true
		}
	}
	if eventID = parseID(chi.URLParam(r, "eventID")); eventID == nil {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	if withComment {
		if commentID = parseID(chi.URLParam(r, "commentID")); commentID == nil {
			writeError(w, http.StatusNotFound, "comment not found")
			return
		}
	}
	ok = true
	return
}

func (h *Handler) listComments(w http.ResponseWriter, r *http.Request) {
	actor, eventID, _, moderator, ok := h.commentContext(w, r, false, false)
	if !ok {
		return
	}
	filter := filterFrom(r)
	page, err := h.service.ListComments(r.Context(), eventID, actor, moderator, filter.Limit, filter.Offset)
	if err != nil {
		h.writeServiceError(w, "list_comments", err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeContent(w, http.StatusOK, "ok", page)
}

type commentBody struct {
	Body string `json:"body"`
}

func (h *Handler) addComment(w http.ResponseWriter, r *http.Request) {
	actor, eventID, _, moderator, ok := h.commentContext(w, r, true, false)
	if !ok {
		return
	}
	var input commentBody
	if err := decodeBody(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	comment, err := h.service.AddComment(r.Context(), eventID, actor, moderator, input.Body)
	if err != nil {
		h.writeServiceError(w, "add_comment", err)
		return
	}
	writeContent(w, http.StatusCreated, "comment posted", comment)
}

func (h *Handler) editComment(w http.ResponseWriter, r *http.Request) {
	actor, eventID, commentID, moderator, ok := h.commentContext(w, r, true, true)
	if !ok {
		return
	}
	var input commentBody
	if err := decodeBody(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	comment, err := h.service.EditComment(r.Context(), eventID, commentID, actor, moderator, input.Body)
	if err != nil {
		h.writeServiceError(w, "edit_comment", err)
		return
	}
	writeContent(w, http.StatusOK, "comment updated", comment)
}

func (h *Handler) deleteComment(w http.ResponseWriter, r *http.Request) {
	actor, eventID, commentID, moderator, ok := h.commentContext(w, r, true, true)
	if !ok {
		return
	}
	if err := h.service.DeleteComment(r.Context(), eventID, commentID, actor, moderator); err != nil {
		h.writeServiceError(w, "delete_comment", err)
		return
	}
	writeContent(w, http.StatusOK, "comment deleted", nil)
}

// callerEvent resolves the caller and event ID from the URL; mutating requests
// must also be same-origin. ok is false once a response has been written.
func (h *Handler) callerEvent(w http.ResponseWriter, r *http.Request, mutating bool) (actor, eventID []byte, moderator, ok bool) {
	if mutating && !h.auth.SameOriginRequest(r) {
		writeError(w, http.StatusForbidden, "request origin is not allowed")
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	actor = parseID(principal.AccountID)
	for _, capability := range principal.Capabilities {
		if capability == auth.CapabilityContentReview {
			moderator = true
		}
	}
	if eventID = parseID(chi.URLParam(r, "eventID")); eventID == nil {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	ok = true
	return
}

func (h *Handler) registrationStatus(w http.ResponseWriter, r *http.Request) {
	actor, eventID, _, ok := h.callerEvent(w, r, false)
	if !ok {
		return
	}
	status, err := h.service.RegistrationStatus(r.Context(), eventID, actor)
	if err != nil {
		h.writeServiceError(w, "registration_status", err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeContent(w, http.StatusOK, "ok", map[string]string{"status": status})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	actor, eventID, _, ok := h.callerEvent(w, r, true)
	if !ok {
		return
	}
	status, err := h.service.Register(r.Context(), eventID, actor)
	if err != nil {
		h.writeServiceError(w, "register_event", err)
		return
	}
	writeContent(w, http.StatusOK, "registered", map[string]string{"status": status})
}

func (h *Handler) unregister(w http.ResponseWriter, r *http.Request) {
	actor, eventID, _, ok := h.callerEvent(w, r, true)
	if !ok {
		return
	}
	if err := h.service.Unregister(r.Context(), eventID, actor); err != nil {
		h.writeServiceError(w, "unregister_event", err)
		return
	}
	writeContent(w, http.StatusOK, "registration cancelled", map[string]string{"status": RegistrationCancelled})
}

func (h *Handler) myRegistrations(w http.ResponseWriter, r *http.Request) {
	principal, _ := auth.PrincipalFromContext(r.Context())
	items, err := h.service.ListMine(r.Context(), parseID(principal.AccountID))
	if err != nil {
		h.fail("list_registrations", err)
		writeError(w, http.StatusInternalServerError, "could not load registrations")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeContent(w, http.StatusOK, "ok", map[string]any{"items": items})
}

func (h *Handler) attendees(w http.ResponseWriter, r *http.Request) {
	actor, eventID, moderator, ok := h.callerEvent(w, r, false)
	if !ok {
		return
	}
	items, err := h.service.ListAttendees(r.Context(), eventID, actor, moderator)
	if err != nil {
		h.writeServiceError(w, "list_attendees", err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeContent(w, http.StatusOK, "ok", map[string]any{"items": items})
}

func (h *Handler) markAttended(w http.ResponseWriter, r *http.Request) {
	actor, eventID, moderator, ok := h.callerEvent(w, r, true)
	if !ok {
		return
	}
	profileID := parseID(chi.URLParam(r, "profileID"))
	if profileID == nil {
		writeError(w, http.StatusNotFound, "registration not found")
		return
	}
	if err := h.service.MarkAttended(r.Context(), eventID, profileID, actor, moderator); err != nil {
		h.writeServiceError(w, "mark_attended", err)
		return
	}
	writeContent(w, http.StatusOK, "attendance recorded", map[string]string{"status": RegistrationAttended})
}
