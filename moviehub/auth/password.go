package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLen = 8
	// bcrypt อ่านรหัสผ่านแค่ 72 ไบต์แรก เกินกว่านั้นถูกตัดทิ้งเงียบ ๆ จึงต้องปฏิเสธตั้งแต่ต้น
	MaxPasswordBytes = 72
)

var (
	ErrPasswordTooShort = errors.New("รหัสผ่านสั้นเกินไป")
	ErrPasswordTooLong  = errors.New("รหัสผ่านยาวเกิน 72 ไบต์")
)

func ValidatePassword(pw string) error {
	if len([]rune(pw)) < MinPasswordLen {
		return ErrPasswordTooShort
	}
	if len(pw) > MaxPasswordBytes {
		return ErrPasswordTooLong
	}
	return nil
}

func HashPassword(pw string) (string, error) {
	if err := ValidatePassword(pw); err != nil {
		return "", err
	}
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// DummyHash ใช้เทียบตอนไม่พบอีเมล เพื่อให้ login ที่อีเมลผิดกับรหัสผิดใช้เวลาเท่ากัน
var DummyHash = func() string {
	h, _ := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)
	return string(h)
}()
