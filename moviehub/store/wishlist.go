package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type WishlistItem struct {
	MovieID int64     `json:"movieId"`
	Note    *string   `json:"note"`
	AddedAt time.Time `json:"addedAt"`
}

func (s *Store) ListWishlist(ctx context.Context, memberID int64) ([]WishlistItem, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT movie_id, note, added_at
         FROM wishlist_items
         WHERE member_id = $1
         ORDER BY added_at DESC, movie_id DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []WishlistItem{}
	for rows.Next() {
		var item WishlistItem
		if err := rows.Scan(&item.MovieID, &item.Note, &item.AddedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) UpsertWishlist(ctx context.Context, memberID, movieID int64, note *string) (WishlistItem, error) {
	var item WishlistItem
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO wishlist_items (member_id, movie_id, note)
         SELECT $1, $2, $3
         WHERE EXISTS (SELECT 1 FROM movies WHERE id = $2)
         ON CONFLICT (member_id, movie_id)
         DO UPDATE SET note = EXCLUDED.note
         RETURNING movie_id, note, added_at`,
		memberID, movieID, note).
		Scan(&item.MovieID, &item.Note, &item.AddedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return WishlistItem{}, ErrNotFound
	}
	return item, err
}

func (s *Store) DeleteWishlist(ctx context.Context, memberID, movieID int64) error {
	_, err := s.Pool.Exec(ctx,
		`DELETE FROM wishlist_items WHERE member_id = $1 AND movie_id = $2`,
		memberID, movieID)
	return err
}
