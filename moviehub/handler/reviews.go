package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"moviehub/mw"
	"moviehub/store"
)

type reviewReq struct {
	Text string `json:"text"`
}

func (h *Handler) AddReview(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	var req reviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	text := strings.TrimSpace(req.Text)
	if utf8.RuneCountInString(text) < 10 {
		writeError(w, http.StatusBadRequest, "REVIEW_TOO_SHORT", "รีวิวสั้นเกินไป")
		return
	}
	memberID, _ := mw.MemberID(r.Context())
	review, err := h.Store.AddReview(r.Context(), memberID, movieID, text)
	switch {
	case errors.Is(err, store.ErrDuplicate):
		writeError(w, http.StatusConflict, "ALREADY_REVIEWED", "คุณรีวิวเรื่องนี้ไปแล้ว")
	case errors.Is(err, store.ErrNoMovie):
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
	case err != nil:
		internal(w, err)
	default:
		writeJSON(w, http.StatusCreated, map[string]any{
			"id": review.ID, "text": review.Text, "createdAt": review.CreatedAt,
		})
	}
}

func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	reviewID, ok := pathID(r, "reviewId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขรีวิวต้องเป็นตัวเลข")
		return
	}
	memberID, _ := mw.MemberID(r.Context())
	err := h.Store.DeleteReview(r.Context(), reviewID, memberID, mw.Role(r.Context()) == "admin")
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "REVIEW_NOT_FOUND", "ไม่พบรีวิวนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
