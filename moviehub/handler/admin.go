package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"moviehub/mw"
	"moviehub/store"
)

// ด่านที่ 5: รายชื่อสมาชิก เฉพาะแอดมิน (route ห่อด้วย RequireRole("admin") แล้ว)
func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.Store.ListMembers(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, members)
}

// ---------------------------------------------------------------------------
// wishlist ของฉัน: member id มาจากตั๋วทุกครั้ง จึงเห็นและแก้ได้เฉพาะของตัวเอง
// ---------------------------------------------------------------------------

func (h *Handler) ListWishlist(w http.ResponseWriter, r *http.Request) {
	c, _ := mw.ClaimsFrom(r.Context())
	items, err := h.Store.ListWishlist(r.Context(), c.Sub)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

type wishlistReq struct {
	Note *string `json:"note"`
}

func (h *Handler) PutWishlist(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(w, r, "movieId")
	if !ok {
		return
	}
	var req wishlistReq
	if r.ContentLength != 0 { // body ว่างได้ = ไม่มีโน้ต
		if !decodeBody(w, r, &req) {
			return
		}
	}
	if req.Note != nil {
		n := strings.TrimSpace(*req.Note)
		if utf8.RuneCountInString(n) > 200 {
			writeError(w, http.StatusBadRequest, "NOTE_TOO_LONG", "โน้ตต้องไม่เกิน 200 ตัวอักษร")
			return
		}
		req.Note = &n
	}

	c, _ := mw.ClaimsFrom(r.Context())
	err := h.Store.PutWishlist(r.Context(), c.Sub, movieID, req.Note)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
	case err != nil:
		serverError(w, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) DeleteWishlist(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(w, r, "movieId")
	if !ok {
		return
	}
	c, _ := mw.ClaimsFrom(r.Context())
	err := h.Store.DeleteWishlist(r.Context(), c.Sub, movieID)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_IN_WISHLIST", "หนังเรื่องนี้ไม่ได้อยู่ในรายการอยากดู")
	case err != nil:
		serverError(w, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
