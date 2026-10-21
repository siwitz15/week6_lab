package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (s *Store) UpsertVote(ctx context.Context, memberID, movieID int64, score int16) error {
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO votes (member_id, movie_id, score)
         SELECT $1, $2, $3
         WHERE EXISTS (SELECT 1 FROM movies WHERE id = $2)
         ON CONFLICT (member_id, movie_id)
         DO UPDATE SET score = EXCLUDED.score, updated_at = now()
         RETURNING score`,
		memberID, movieID, score,
	).Scan(&score)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoMovie
	}
	return err
}
