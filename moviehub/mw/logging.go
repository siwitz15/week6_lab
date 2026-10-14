package mw

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// ไม่ log header Authorization และไม่ log query ทั้งก้อน เพื่อไม่ให้ตั๋วหลุดลง log
		log.Printf("%s %s ใช้เวลา %v", r.Method, r.URL.Path,
			time.Since(start).Round(time.Millisecond))
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    "INTERNAL_ERROR",
						"message": "ระบบมีปัญหา ลองใหม่อีกครั้ง",
					},
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
