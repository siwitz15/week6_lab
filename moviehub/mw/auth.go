package mw

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"moviehub/auth"
)

type ctxKey struct{}

// ClaimsFrom ดึงข้อมูลจากตั๋วที่ผ่านการตรวจแล้ว ใช้เป็นแหล่งเดียวของ "ผู้ใช้คนนี้คือใคร"
func ClaimsFrom(ctx context.Context) (auth.Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(auth.Claims)
	return c, ok
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": msg},
	})
}

// Auth ตรวจตั๋วจาก Authorization: Bearer <token>
// ทุกเหตุผลที่ไม่ผ่านตอบ 401 ข้อความเดียวกัน ไม่บอกว่าผิดที่ด่านไหน
func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "ต้องเข้าสู่ระบบก่อน")
				return
			}
			claims, err := auth.Verify(secret, token, time.Now())
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "ต้องเข้าสู่ระบบก่อน")
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole ใช้ต่อท้าย Auth ตั๋วถูกต้องแต่บทบาทไม่พอ = 403
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, ok := ClaimsFrom(r.Context())
			if !ok {
				writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "ต้องเข้าสู่ระบบก่อน")
				return
			}
			if c.Role != role {
				writeErr(w, http.StatusForbidden, "FORBIDDEN", "ไม่มีสิทธิ์ใช้งานส่วนนี้")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
