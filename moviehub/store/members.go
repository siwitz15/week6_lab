package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Member struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

// MemberWithHash ใช้ภายในตอน login เท่านั้น ห้ามส่งออกเป็น JSON
type MemberWithHash struct {
	Member
	PasswordHash string
}

func (s *Store) CreateMember(ctx context.Context, email, hash, displayName string) (Member, error) {
	var m Member
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO members (email, password_hash, display_name)
		VALUES ($1, $2, $3)
		RETURNING id, email, display_name, role, created_at`,
		strings.ToLower(email), hash, displayName,
	).Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt)
	if isUnique(err) {
		return Member{}, ErrDuplicate
	}
	return m, err
}

func (s *Store) FindMemberByEmail(ctx context.Context, email string) (MemberWithHash, error) {
	var m MemberWithHash
	err := s.Pool.QueryRow(ctx, `
		SELECT id, email, display_name, role, created_at, password_hash
		FROM members WHERE lower(email) = lower($1)`, email,
	).Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt, &m.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return MemberWithHash{}, ErrNotFound
	}
	return m, err
}

func (s *Store) GetMember(ctx context.Context, id int64) (Member, error) {
	var m Member
	err := s.Pool.QueryRow(ctx, `
		SELECT id, email, display_name, role, created_at
		FROM members WHERE id = $1`, id,
	).Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	return m, err
}

// ListMembers สำหรับหน้าแอดมิน ไม่ดึง password_hash มาเลย
func (s *Store) ListMembers(ctx context.Context) ([]Member, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, email, display_name, role, created_at
		FROM members ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []Member{}
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// ---------------------------------------------------------------------------
// wishlist: ทุกคำสั่งผูกกับ member_id จากตั๋ว จึงแตะรายการของคนอื่นไม่ได้
// ---------------------------------------------------------------------------

type WishlistItem struct {
	MovieID int64     `json:"movie_id"`
	Title   string    `json:"title"`
	Note    *string   `json:"note"`
	AddedAt time.Time `json:"added_at"`
}

func (s *Store) ListWishlist(ctx context.Context, memberID int64) ([]WishlistItem, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT w.movie_id, m.title, w.note, w.added_at
		FROM wishlist_items w JOIN movies m ON m.id = w.movie_id
		WHERE w.member_id = $1
		ORDER BY w.added_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []WishlistItem{}
	for rows.Next() {
		var it WishlistItem
		if err := rows.Scan(&it.MovieID, &it.Title, &it.Note, &it.AddedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// PutWishlist เพิ่มหรือแก้โน้ต (upsert) ไม่พบหนัง = ErrNotFound
func (s *Store) PutWishlist(ctx context.Context, memberID, movieID int64, note *string) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO wishlist_items (member_id, movie_id, note)
		VALUES ($1, $2, $3)
		ON CONFLICT (member_id, movie_id) DO UPDATE SET note = EXCLUDED.note`,
		memberID, movieID, note)
	if isFK(err) {
		return ErrNotFound
	}
	return err
}

func (s *Store) DeleteWishlist(ctx context.Context, memberID, movieID int64) error {
	tag, err := s.Pool.Exec(ctx, `
		DELETE FROM wishlist_items WHERE member_id = $1 AND movie_id = $2`,
		memberID, movieID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
