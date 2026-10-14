package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"moviehub/mw"
	"moviehub/store"
)

type Handler struct {
	Store     *store.Store
	JWTSecret []byte
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("store error: %v", err)
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "ระบบมีปัญหา ลองใหม่อีกครั้ง")
}

// decodeBody อ่าน JSON ขนาดไม่เกิน 8 KB และไม่รับ field แปลกปลอม
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านข้อมูล JSON ไม่ออก")
		return false
	}
	return true
}

func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "BAD_ID", "id ต้องเป็นตัวเลข")
		return 0, false
	}
	return id, true
}

// ด่านที่ 0
func (h *Handler) ListMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := h.Store.SearchMovies(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movies)
}

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	reviews, err := h.Store.ListReviews(r.Context(), id)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reviews)
}

// Routes: ใครเปิดได้บ้างดูจากการห่อ middleware ที่แต่ละบรรทัด
func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	authed := mw.Auth(h.JWTSecret)
	adminOnly := func(f http.HandlerFunc) http.Handler {
		return authed(mw.RequireRole("admin")(f))
	}
	member := func(f http.HandlerFunc) http.Handler { return authed(f) }

	// สาธารณะ
	mux.HandleFunc("GET /api/movies", h.ListMovies)
	mux.HandleFunc("GET /api/movies/{id}/reviews", h.ListReviews)
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)

	// ต้องมีตั๋ว
	mux.Handle("GET /api/me", member(h.Me))
	mux.Handle("POST /api/movies/{id}/reviews", member(h.CreateReview))
	mux.Handle("DELETE /api/reviews/{reviewId}", member(h.DeleteReview))
	mux.Handle("GET /api/me/wishlist", member(h.ListWishlist))
	mux.Handle("PUT /api/me/wishlist/{movieId}", member(h.PutWishlist))
	mux.Handle("DELETE /api/me/wishlist/{movieId}", member(h.DeleteWishlist))

	// เฉพาะแอดมิน
	mux.Handle("GET /api/admin/members", adminOnly(h.ListMembers))
	return mux
}
