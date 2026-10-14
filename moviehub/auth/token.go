package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Claims คือข้อมูลที่เขียนบนตั๋ว (ใครเป็นเจ้าของ บทบาทอะไร หมดอายุเมื่อไร)
type Claims struct {
	Sub  int64  `json:"sub"`  // member id
	Role string `json:"role"` // member | admin
	Iat  int64  `json:"iat"`
	Exp  int64  `json:"exp"`
}

// เหตุผลที่ตั๋วถูกปฏิเสธ แต่ละข้อตรงกับด่านตรวจใน Verify
var (
	ErrMalformed = errors.New("ตั๋วรูปแบบผิด")
	ErrAlg       = errors.New("วิธีเซ็นไม่ใช่ HS256")
	ErrSignature = errors.New("ลายเซ็นไม่ตรง")
	ErrExpired   = errors.New("ตั๋วหมดอายุ")
	ErrClaims    = errors.New("ข้อมูลบนตั๋วไม่ครบหรือไม่ถูกต้อง")
)

var enc = base64.RawURLEncoding

// header คงที่ ใช้ทั้งตอนออกและตอนตรวจ
const headerJSON = `{"alg":"HS256","typ":"JWT"}`

func sign(secret []byte, signingInput string) []byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(signingInput))
	return m.Sum(nil)
}

// Sign ออกตั๋วใหม่ อายุ ttl นับจาก now
func Sign(secret []byte, memberID int64, role string, now time.Time, ttl time.Duration) (string, error) {
	payload, err := json.Marshal(Claims{
		Sub:  memberID,
		Role: role,
		Iat:  now.Unix(),
		Exp:  now.Add(ttl).Unix(),
	})
	if err != nil {
		return "", err
	}
	input := enc.EncodeToString([]byte(headerJSON)) + "." + enc.EncodeToString(payload)
	return input + "." + enc.EncodeToString(sign(secret, input)), nil
}

// Verify ตรวจตั๋วตามลำดับ 5 ด่าน ด่านไหนไม่ผ่านคืน error ของด่านนั้นทันที
func Verify(secret []byte, token string, now time.Time) (Claims, error) {
	var zero Claims

	// ด่าน 1: ต้องมี 3 ท่อน คั่นด้วยจุด
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return zero, ErrMalformed
	}

	// ด่าน 2: header ต้องอ่านได้ และ alg ต้องเป็น HS256 เท่านั้น
	// (ปฏิเสธ "none" และ alg อื่น ไม่ให้ผู้ส่งเลือกวิธีตรวจเอง)
	rawHeader, err := enc.DecodeString(parts[0])
	if err != nil {
		return zero, ErrMalformed
	}
	var h struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(rawHeader, &h); err != nil {
		return zero, ErrMalformed
	}
	if h.Alg != "HS256" {
		return zero, ErrAlg
	}

	// ด่าน 3: ลายเซ็นต้องตรง เทียบแบบ constant-time
	gotSig, err := enc.DecodeString(parts[2])
	if err != nil {
		return zero, ErrMalformed
	}
	if !hmac.Equal(gotSig, sign(secret, parts[0]+"."+parts[1])) {
		return zero, ErrSignature
	}

	// ด่าน 4: อ่านเนื้อความ (เชื่อได้แล้วเพราะลายเซ็นผ่าน)
	rawPayload, err := enc.DecodeString(parts[1])
	if err != nil {
		return zero, ErrMalformed
	}
	var c Claims
	if err := json.Unmarshal(rawPayload, &c); err != nil {
		return zero, ErrMalformed
	}

	// ด่าน 5: ต้องไม่หมดอายุ และข้อมูลต้องครบ
	if c.Exp == 0 || now.Unix() >= c.Exp {
		return zero, ErrExpired
	}
	if c.Sub <= 0 || (c.Role != "member" && c.Role != "admin") {
		return zero, ErrClaims
	}
	return c, nil
}

// Peek อ่านเนื้อความของตั๋วโดยไม่ตรวจลายเซ็นและไม่ใช้กุญแจ
// (ตั๋วแค่เข้ารหัส base64 ไม่ได้ซ่อนข้อมูล ห้ามใส่ความลับลงตั๋ว)
func Peek(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrMalformed
	}
	raw, err := enc.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrMalformed
	}
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return Claims{}, ErrMalformed
	}
	return c, nil
}
