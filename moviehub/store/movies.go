package store

import (
	"context"
	"strings"
	"time"
)

type Movie struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	TitleTH     *string    `json:"title_th"`
	Overview    *string    `json:"overview"`
	PosterPath  *string    `json:"poster_path"`
	ReleaseDate *time.Time `json:"release_date"`
	VoteAverage float64    `json:"vote_average"`
}

// escapeLike กัน % และ _ ที่ผู้ใช้พิมพ์มาไม่ให้กลายเป็นตัวแทนอักขระ
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func (s *Store) SearchMovies(ctx context.Context, q string) ([]Movie, error) {
	pattern := "%" + escapeLike(strings.ToLower(strings.TrimSpace(q))) + "%"
	rows, err := s.Pool.Query(ctx, `
		SELECT id, title, title_th, overview, poster_path, release_date, vote_average::float8
		FROM movies
		WHERE lower(title) LIKE $1 ESCAPE '\'
		   OR lower(coalesce(title_th, '')) LIKE $1 ESCAPE '\'
		ORDER BY popularity DESC
		LIMIT 100`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := []Movie{}
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.TitleTH, &m.Overview,
			&m.PosterPath, &m.ReleaseDate, &m.VoteAverage); err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, rows.Err()
}
