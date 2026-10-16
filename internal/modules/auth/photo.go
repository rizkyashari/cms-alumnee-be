package auth

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// servePhoto writes a stored profile photo. The bytes were re-encoded as
// JPEG/PNG at upload, but the response still forbids sniffing and sandboxes
// the content so it can only ever render as an image.
func servePhoto(w http.ResponseWriter, mimeType string, data []byte) {
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("Cache-Control", "private, no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ownPhoto returns the authenticated member's own profile photo.
func (h *Handler) ownPhoto(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var mimeType string
	var data []byte
	err := h.service.db.QueryRowContext(r.Context(), `
		SELECT mime_type, image_data FROM alumni_profile_photos WHERE account_id = ?`,
		principalIDBytes(principal.AccountID)).Scan(&mimeType, &data)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "photo not found")
		return
	}
	if err != nil {
		h.reportFailure("read_own_photo", err)
		writeError(w, http.StatusInternalServerError, "could not load photo")
		return
	}
	servePhoto(w, mimeType, data)
}

// reviewPhoto returns the photo of a profile with a pending verification
// request. It is limited to Pengurus/Admin and verified members of the same
// batch, who need the photo to recognise the person they are vouching for.
func (h *Handler) reviewPhoto(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	profileID := principalIDBytes(chi.URLParam(r, "profileID"))
	if profileID == nil {
		writeError(w, http.StatusNotFound, "photo not found")
		return
	}
	var mimeType string
	var data []byte
	err := h.service.db.QueryRowContext(r.Context(), `
		SELECT p.mime_type, p.image_data
		FROM alumni_profiles ap
		JOIN alumni_profile_photos p ON p.account_id = ap.account_id
		JOIN verification_requests vr ON vr.profile_id = ap.id AND vr.status = 'pending'
		WHERE ap.id = ?
			AND (? <> 'verified' OR ap.batch_year = (
				SELECT me.batch_year FROM alumni_profiles me WHERE me.account_id = ?))`,
		profileID, string(principal.Role), principalIDBytes(principal.AccountID)).Scan(&mimeType, &data)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "photo not found")
		return
	}
	if err != nil {
		h.reportFailure("read_review_photo", err)
		writeError(w, http.StatusInternalServerError, "could not load photo")
		return
	}
	servePhoto(w, mimeType, data)
}
