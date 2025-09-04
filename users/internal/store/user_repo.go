package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserModel struct {
	ID            int64
	Email         string
	PasswordHash  string
	Role          string
	Nick          string
	AvatarURL     *string
	Bio           *string
	Links         []string
	IsBanned      bool
	EmailVerified bool
	CreatedAt     time.Time
}

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) CreateUser(ctx context.Context, email, passHash, nick string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
	  insert into users (email, password_hash, nick)
	  values (lower($1), $2, $3)
	  returning id
	`, email, passHash, nick).Scan(&id)
	return id, err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*UserModel, error) {
	row := r.db.QueryRow(ctx, `
	  select id, email, password_hash, role, nick, avatar_url, bio, links, is_banned, email_verified, created_at
	  from users where email = lower($1)
	`, email)
	u := UserModel{}
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Nick, &u.AvatarURL, &u.Bio, &u.Links, &u.IsBanned, &u.EmailVerified, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) VerifyEmail(ctx context.Context, token string) (bool, error) {
	ct, err := r.db.Exec(ctx, `
	  update users set email_verified = true
	  where id = (select user_id from email_verify_tokens where token=$1 and expires_at > now())
	`, token)
	if err != nil {
		return false, err
	}
	if ct.RowsAffected() == 0 {
		return false, nil
	}
	_, _ = r.db.Exec(ctx, `delete from email_verify_tokens where token=$1`, token)
	return true, nil
}

func (r *UserRepo) UpsertVerifyToken(ctx context.Context, userID int64, token string, ttl time.Duration) error {
	_, err := r.db.Exec(ctx, `
	  insert into email_verify_tokens (token, user_id, expires_at)
	  values ($1, $2, now() + $3::interval)
	`, token, userID, ttl.String())
	return err
}

func (r *UserRepo) UpdateProfile(ctx context.Context, userID int64, nick string, avatar *string, bio *string, links []string) error {
	_, err := r.db.Exec(ctx, `
	  update users
	  set nick=$2, avatar_url=$3, bio=$4, links=$5
	  where id=$1
	`, userID, nick, avatar, bio, links)
	return err
}

// Пагинация сабмитов (пока демо-данные из jam_submissions)
type SubmissionItem struct {
	Jam         string
	Game        string
	Status      string
	SubmittedAt time.Time
}

func (r *UserRepo) ListMySubmissions(ctx context.Context, userID int64, page, limit int) ([]SubmissionItem, int, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	rows, err := r.db.Query(ctx, `
	  select jam, game, status, submitted_at
	  from jam_submissions
	  where user_id=$1
	  order by submitted_at desc
	  limit $2 offset $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []SubmissionItem{}
	for rows.Next() {
		var it SubmissionItem
		if err := rows.Scan(&it.Jam, &it.Game, &it.Status, &it.SubmittedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}
	var total int
	_ = r.db.QueryRow(ctx, `select count(*) from jam_submissions where user_id=$1`, userID).Scan(&total)
	return items, total, nil
}
