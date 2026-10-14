package handler

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"moviehub/auth"
	"moviehub/mw"
	"moviehub/store"
)

const tokenTTL = time.Hour

type registerReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// ด่านที่ 1: สมัครสมาชิก
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if !decodeBody(w, r, &req) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.DisplayName)

	if _, err := mail.ParseAddress(email); err != nil || len(email) > 255 {
		writeError(w, http.StatusBadRequest, "BAD_EMAIL", "รูปแบบอีเมลไม่ถูกต้อง")
		return
	}
	if n := utf8.RuneCountInString(name); n < 1 || n > 50 {
		writeError(w, http.StatusBadRequest, "BAD_NAME", "ชื่อที่แสดงต้องยาว 1-50 ตัวอักษร")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	switch {
	case errors.Is(err, auth.ErrPasswordTooShort):
		writeError(w, http.StatusBadRequest, "PASSWORD_TOO_SHORT", "รหัสผ่านต้องยาวอย่างน้อย 8 ตัวอักษร")
		return
	case errors.Is(err, auth.ErrPasswordTooLong):
		writeError(w, http.StatusBadRequest, "PASSWORD_TOO_LONG", "รหัสผ่านต้องไม่เกิน 72 ไบต์")
		return
	case err != nil:
		serverError(w, err)
		return
	}

	m, err := h.Store.CreateMember(r.Context(), email, hash, name)
	if errors.Is(err, store.ErrDuplicate) {
		writeError(w, http.StatusConflict, "EMAIL_TAKEN", "อีเมลนี้ถูกใช้แล้ว")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ด่านที่ 2 และ 7: เข้าสู่ระบบ
// อีเมลผิดกับรหัสผิดต้องตอบเหมือนกันทุกอย่าง ทั้งข้อความและเวลา
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if !decodeBody(w, r, &req) {
		return
	}
	email := strings.TrimSpace(req.Email)

	m, err := h.Store.FindMemberByEmail(r.Context(), email)
	hash := auth.DummyHash // ไม่พบอีเมลก็ยังเทียบ bcrypt กับแฮชหลอก ให้ใช้เวลาใกล้เคียงกัน
	if err == nil {
		hash = m.PasswordHash
	} else if !errors.Is(err, store.ErrNotFound) {
		serverError(w, err)
		return
	}
	passOK := auth.CheckPassword(hash, req.Password)

	if err != nil || !passOK {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	token, err := auth.Sign(h.JWTSecret, m.ID, m.Role, time.Now(), tokenTTL)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "member": m.Member})
}

// ด่านที่ 4: ฉันคือใคร (อ่านจากตั๋ว แล้วไปดึงข้อมูลสด)
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	c, _ := mw.ClaimsFrom(r.Context())
	m, err := h.Store.GetMember(r.Context(), c.Sub)
	if errors.Is(err, store.ErrNotFound) {
		// ตั๋วถูกต้องแต่สมาชิกถูกลบไปแล้ว
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "ต้องเข้าสู่ระบบก่อน")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}
