package store

import (
	"context"
	"time"
)

type Review struct {
	ID          int64     `json:"id"`
	MovieID     int64     `json:"movie_id"`
	MemberID    int64     `json:"member_id"`
	DisplayName string    `json:"display_name"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Store) ListReviews(ctx context.Context, movieID int64) ([]Review, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT r.id, r.movie_id, r.member_id, m.display_name, r.text, r.created_at
		FROM reviews r JOIN members m ON m.id = r.member_id
		WHERE r.movie_id = $1
		ORDER BY r.created_at DESC`, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []Review{}
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.MovieID, &r.MemberID, &r.DisplayName, &r.Text, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

// CreateReview: ไม่พบหนัง = ErrNotFound, เคยรีวิวเรื่องนี้แล้ว = ErrDuplicate
func (s *Store) CreateReview(ctx context.Context, memberID, movieID int64, text string) (Review, error) {
	var r Review
	err := s.Pool.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO reviews (member_id, movie_id, text)
			VALUES ($1, $2, $3)
			RETURNING id, movie_id, member_id, text, created_at
		)
		SELECT ins.id, ins.movie_id, ins.member_id, m.display_name, ins.text, ins.created_at
		FROM ins JOIN members m ON m.id = ins.member_id`,
		memberID, movieID, text,
	).Scan(&r.ID, &r.MovieID, &r.MemberID, &r.DisplayName, &r.Text, &r.CreatedAt)
	switch {
	case isUnique(err):
		return Review{}, ErrDuplicate
	case isFK(err):
		return Review{}, ErrNotFound
	}
	return r, err
}

// DeleteReview ลบได้เมื่อเป็นเจ้าของหรือแอดมิน
// ไม่ใช่เจ้าของก็ตอบ ErrNotFound เหมือนไม่มีรีวิวนั้น เพื่อไม่บอกว่ามีรีวิวของคนอื่นอยู่
func (s *Store) DeleteReview(ctx context.Context, reviewID, memberID int64, isAdmin bool) error {
	tag, err := s.Pool.Exec(ctx, `
		DELETE FROM reviews
		WHERE id = $1 AND (member_id = $2 OR $3)`,
		reviewID, memberID, isAdmin)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
