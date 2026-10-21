package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"moviehub/auth"
	"moviehub/mw"
)

func TestVoteMovieRejectsInvalidScores(t *testing.T) {
	issuer := &auth.Issuer{Secret: []byte("test-secret-that-is-long-enough"), TTL: time.Hour}
	token, err := issuer.Issue(1, "member")
	if err != nil {
		t.Fatal(err)
	}

	handler := mw.RequireLogin(issuer, (&Handler{}).VoteMovie)
	for _, body := range []string{
		`{"score":0}`,
		`{"score":11}`,
		`{"score":8.5}`,
		`{"score":"8"}`,
		`{}`,
		`{"score":8} {"score":9}`,
	} {
		t.Run(body, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/api/movies/1621552/vote", strings.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.SetPathValue("id", "1621552")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", got)
			}
			var response struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatalf("response is not JSON: %v", err)
			}
			if response.Error != "score ต้องเป็นจำนวนเต็ม 1 ถึง 10" {
				t.Fatalf("error = %q", response.Error)
			}
		})
	}
}

func TestVoteMovieRequiresLogin(t *testing.T) {
	issuer := &auth.Issuer{Secret: []byte("test-secret-that-is-long-enough"), TTL: time.Hour}
	handler := mw.RequireLogin(issuer, (&Handler{}).VoteMovie)
	req := httptest.NewRequest(http.MethodPut, "/api/movies/1621552/vote", strings.NewReader(`{"score":8}`))
	req.SetPathValue("id", "1621552")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
}
