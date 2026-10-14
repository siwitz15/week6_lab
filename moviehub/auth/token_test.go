package auth

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

var testSecret = []byte("0123456789abcdef0123456789abcdef")

func forge(t *testing.T, header string, c Claims, secret []byte) string {
	t.Helper()
	payload, _ := json.Marshal(c)
	input := enc.EncodeToString([]byte(header)) + "." + enc.EncodeToString(payload)
	return input + "." + enc.EncodeToString(sign(secret, input))
}

func TestVerifyGoodTicket(t *testing.T) {
	now := time.Now()
	tok, err := Sign(testSecret, 7, "member", now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	c, err := Verify(testSecret, tok, now)
	if err != nil {
		t.Fatalf("ตั๋วจริงต้องผ่าน: %v", err)
	}
	if c.Sub != 7 || c.Role != "member" {
		t.Fatalf("เนื้อความผิด: %+v", c)
	}
}

// ตั๋วปลอม 6 แบบ ต้องถูกปฏิเสธทั้งหมด และถูกปฏิเสธที่ด่านที่คาดไว้
func TestVerifyForgedTickets(t *testing.T) {
	now := time.Now()
	good := Claims{Sub: 7, Role: "member", Iat: now.Unix(), Exp: now.Add(time.Hour).Unix()}
	goodTok, _ := Sign(testSecret, 7, "member", now, time.Hour)

	// แบบที่ 4: แก้ role เป็น admin แต่ใช้ลายเซ็นเดิม
	adminPayload, _ := json.Marshal(Claims{Sub: 7, Role: "admin", Iat: good.Iat, Exp: good.Exp})
	tampered := goodTok[:len(enc.EncodeToString([]byte(headerJSON)))+1] +
		enc.EncodeToString(adminPayload) + goodTok[lastDot(goodTok):]

	cases := []struct {
		name  string
		token string
		want  error
	}{
		{"1 ไม่ใช่รูปแบบตั๋ว", "not-a-token", ErrMalformed},
		{"2 alg เป็น none ไม่มีลายเซ็น",
			enc.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`)) + "." +
				enc.EncodeToString([]byte(`{"sub":1,"role":"admin","exp":99999999999}`)) + ".", ErrAlg},
		{"3 alg เป็น HS512 ไม่ตรงที่ร้านใช้",
			forge(t, `{"alg":"HS512","typ":"JWT"}`, good, testSecret), ErrAlg},
		{"4 แก้เนื้อความเป็น admin แต่ใช้ลายเซ็นเดิม", tampered, ErrSignature},
		{"5 เซ็นด้วยกุญแจอื่น",
			forge(t, headerJSON, good, []byte("another-secret-another-secret-123456")), ErrSignature},
		{"6 ตั๋วหมดอายุ",
			forge(t, headerJSON, Claims{Sub: 7, Role: "member", Iat: now.Add(-2 * time.Hour).Unix(), Exp: now.Add(-time.Hour).Unix()}, testSecret), ErrExpired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Verify(testSecret, tc.token, now)
			if !errors.Is(err, tc.want) {
				t.Fatalf("ต้องถูกปฏิเสธด้วย %v แต่ได้ %v", tc.want, err)
			}
		})
	}
}

func TestPeekReadsWithoutKey(t *testing.T) {
	tok, _ := Sign(testSecret, 42, "admin", time.Now(), time.Hour)
	c, err := Peek(tok)
	if err != nil || c.Sub != 42 || c.Role != "admin" {
		t.Fatalf("Peek ผิด: %+v %v", c, err)
	}
}

func lastDot(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}
