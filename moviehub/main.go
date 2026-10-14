package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"moviehub/handler"
	"moviehub/mw"
	"moviehub/store"
)

// loadSecret อ่านกุญแจเซ็นตั๋วจาก JWT_SECRET ถ้าไม่มีลองอ่านไฟล์ .jwt_secret
// ต้องยาวอย่างน้อย 32 ตัวอักษร ไม่งั้นไม่ยอมเปิดเซิร์ฟเวอร์
func loadSecret() []byte {
	s := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if s == "" {
		if b, err := os.ReadFile(".jwt_secret"); err == nil {
			s = strings.TrimSpace(string(b))
		}
	}
	if len(s) < 32 {
		log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร (ตั้งตัวแปรสภาพแวดล้อม หรือสร้างไฟล์ .jwt_secret)")
	}
	return []byte(s)
}

func main() {
	ctx := context.Background()

	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("ต่อฐานข้อมูลไม่ติด ", err)
	}

	origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	h := &handler.Handler{Store: &store.Store{Pool: pool}, JWTSecret: loadSecret()}

	var app http.Handler = http.TimeoutHandler(h.Routes(), 10*time.Second, "ใช้เวลานานเกินไป")
	// CORS อยู่นอกสุดของ handler จริง เพื่อให้ preflight จบก่อนถึงด่านตรวจตั๋ว
	app = mw.Logging(mw.Recovery(mw.CORS(origins)(app)))

	srv := &http.Server{Addr: ":8080", Handler: app}
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	log.Println("moviehub เปิดที่ :8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("ปิดไม่ทัน", err)
	}
	log.Println("ปิดเซิร์ฟเวอร์เรียบร้อย")
}
