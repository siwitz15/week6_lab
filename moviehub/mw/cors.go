package mw

import (
	"net/http"
	"strings"
)

// CORS อนุญาตเฉพาะ origin ที่อยู่ในรายการ (ไม่ใช้ "*" และไม่สะท้อน Origin ที่ส่งมาแบบไม่ตรวจ)
// allowed คือรายการเต็ม เช่น "http://localhost:5173"
func CORS(allowed []string) func(http.Handler) http.Handler {
	set := make(map[string]bool, len(allowed))
	for _, o := range allowed {
		if o = strings.TrimSpace(o); o != "" {
			set[o] = true
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			w.Header().Add("Vary", "Origin")

			if origin != "" && set[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// preflight: จบที่นี่ ไม่เข้า handler จริง
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				if origin != "" && set[origin] {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					w.Header().Set("Access-Control-Max-Age", "600")
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
