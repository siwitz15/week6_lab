package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"moviehub/mw"
	"moviehub/store"
)

type wishlistReq struct {
	Note string `json:"note"`
}

func (h *Handler) ListWishlist(w http.ResponseWriter, r *http.Request) {
	memberID, _ := mw.MemberID(r.Context())
	items, err := h.Store.ListWishlist(r.Context(), memberID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) UpsertWishlist(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "movieId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	var req wishlistReq
	if r.Body != http.NoBody {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if !errors.Is(err, io.EOF) {
				writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
				return
			}
		}
	}
	req.Note = strings.TrimSpace(req.Note)
	if utf8.RuneCountInString(req.Note) > 200 {
		writeError(w, http.StatusBadRequest, "BAD_NOTE", "โน้ตยาวเกินไป")
		return
	}
	var note *string
	if req.Note != "" {
		note = &req.Note
	}
	memberID, _ := mw.MemberID(r.Context())
	item, err := h.Store.UpsertWishlist(r.Context(), memberID, movieID, note)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteWishlist(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "movieId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	memberID, _ := mw.MemberID(r.Context())
	if err := h.Store.DeleteWishlist(r.Context(), memberID, movieID); err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
