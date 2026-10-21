package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"moviehub/mw"
	"moviehub/store"
)

type voteReq struct {
	Score int16 `json:"score"`
}

func (h *Handler) VoteMovie(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}

	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}

	var req voteReq
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "score ต้องเป็นจำนวนเต็ม 1 ถึง 10",
		})
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "score ต้องเป็นจำนวนเต็ม 1 ถึง 10",
		})
		return
	}
	if req.Score < 1 || req.Score > 10 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "score ต้องเป็นจำนวนเต็ม 1 ถึง 10",
		})
		return
	}

	if err := h.Store.UpsertVote(r.Context(), memberID, movieID, req.Score); errors.Is(err, store.ErrNoMovie) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	} else if err != nil {
		internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"movieId": movieID,
		"score":   req.Score,
	})
}
