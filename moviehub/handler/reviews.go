package handler

import (
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

// ด่านที่ 4: เขียนรีวิว
// ผู้เขียนมาจากตั๋วเท่านั้น struct ไม่มี field member_id และ DisallowUnknownFields
// ทำให้ถ้า client แอบส่ง member_id มาจะถูกปฏิเสธ 400
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var req reviewReq
	if !decodeBody(w, r, &req) {
		return
	}
	text := strings.TrimSpace(req.Text)
	if utf8.RuneCountInString(text) < 10 {
		writeError(w, http.StatusBadRequest, "REVIEW_TOO_SHORT", "รีวิวต้องยาวอย่างน้อย 10 ตัวอักษร")
		return
	}

	c, _ := mw.ClaimsFrom(r.Context())
	rev, err := h.Store.CreateReview(r.Context(), c.Sub, movieID, text)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
	case errors.Is(err, store.ErrDuplicate):
		writeError(w, http.StatusConflict, "ALREADY_REVIEWED", "คุณรีวิวหนังเรื่องนี้แล้ว")
	case err != nil:
		serverError(w, err)
	default:
		writeJSON(w, http.StatusCreated, rev)
	}
}

// ด่านที่ 5: ลบรีวิว (เจ้าของหรือแอดมิน)
func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	reviewID, ok := pathID(w, r, "reviewId")
	if !ok {
		return
	}
	c, _ := mw.ClaimsFrom(r.Context())
	err := h.Store.DeleteReview(r.Context(), reviewID, c.Sub, c.Role == "admin")
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "REVIEW_NOT_FOUND", "ไม่พบรีวิวนี้")
	case err != nil:
		serverError(w, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
