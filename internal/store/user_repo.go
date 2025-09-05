package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/RESERPIX/hubigr/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// CreateUser - UC-1.1.1
func (r *UserRepo) CreateUser(ctx context.Context, req domain.SignUpRequest, hash string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email, hash, nick, role) 
		VALUES (LOWER($1), $2, $3, $4) 
		RETURNING id`,
		req.Email, hash, req.Nick, domain.RoleParticipant).Scan(&id)
	return id, err
}

// GetByEmail - UC-1.1.2
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	var linksJSON []byte
	var privacyJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, email, hash, role, nick, avatar, bio, links, is_banned, 
		       email_verified, created_at, privacy_settings
		FROM users WHERE email = LOWER($1)`, email).Scan(
		&u.ID, &u.Email, &u.Hash, &u.Role, &u.Nick, &u.Avatar, &u.Bio,
		&linksJSON, &u.IsBanned, &u.EmailVerified, &u.CreatedAt, &privacyJSON)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(linksJSON, &u.Links); err != nil {
		u.Links = []domain.Link{}
	}
	if err := json.Unmarshal(privacyJSON, &u.PrivacySettings); err != nil {
		u.PrivacySettings = domain.PrivacySettings{}
	}
	return &u, nil
}

// CreateVerifyToken - UC-1.1.1 (TTL 1 час)
func (r *UserRepo) CreateVerifyToken(ctx context.Context, userID int64, token string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO email_verify_tokens (token, user_id, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '1 hour')
		ON CONFLICT (user_id) DO UPDATE SET token = $1, expires_at = NOW() + INTERVAL '1 hour'`,
		token, userID)
	return err
}

// VerifyEmail - UC-1.1.1
func (r *UserRepo) VerifyEmail(ctx context.Context, token string) (bool, error) {
	result, err := r.db.Exec(ctx, `
		UPDATE users SET email_verified = true
		WHERE id = (
			SELECT user_id FROM email_verify_tokens 
			WHERE token = $1 AND expires_at > NOW()
		)`, token)

	if err != nil {
		return false, err
	}

	if result.RowsAffected() == 0 {
		return false, nil
	}

	// Удаляем использованный токен
	r.db.Exec(ctx, "DELETE FROM email_verify_tokens WHERE token = $1", token)
	return true, nil
}

// UpdateProfile - UC-1.2.1
func (r *UserRepo) UpdateProfile(ctx context.Context, userID int64, req domain.UpdateProfileRequest) error {
	linksJSON, err := json.Marshal(req.Links)
	if err != nil {
		return err
	}
	privacyJSON, err2 := json.Marshal(req.PrivacySettings)
	if err2 != nil {
		return err2
	}

	_, err3 := r.db.Exec(ctx, `
		UPDATE users 
		SET nick = $2, avatar = $3, bio = $4, links = $5, privacy_settings = $6
		WHERE id = $1`,
		userID, req.Nick, req.Avatar, req.Bio, linksJSON, privacyJSON)
	return err3
}

// GetByID - для профиля
func (r *UserRepo) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	var u domain.User
	var linksJSON []byte
	var privacyJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, email, hash, role, nick, avatar, bio, links, is_banned,
		       email_verified, created_at, privacy_settings
		FROM users WHERE id = $1`, userID).Scan(
		&u.ID, &u.Email, &u.Hash, &u.Role, &u.Nick, &u.Avatar, &u.Bio,
		&linksJSON, &u.IsBanned, &u.EmailVerified, &u.CreatedAt, &privacyJSON)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(linksJSON, &u.Links); err != nil {
		u.Links = []domain.Link{}
	}
	if err := json.Unmarshal(privacyJSON, &u.PrivacySettings); err != nil {
		u.PrivacySettings = domain.PrivacySettings{}
	}
	return &u, nil
}

// UpdateNotificationSettings - UC-1.2.3
func (r *UserRepo) UpdateNotificationSettings(ctx context.Context, userID int64, settings domain.NotificationSettings) error {
	channelsJSON, err := json.Marshal(settings.Channels)
	if err != nil {
		return err
	}
	_, err2 := r.db.Exec(ctx, `
		INSERT INTO notification_settings (user_id, new_game, new_build, new_post, channels)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			new_game = $2, new_build = $3, new_post = $4, channels = $5`,
		userID, settings.NewGame, settings.NewBuild, settings.NewPost, channelsJSON)
	return err2
}

// GetNotificationSettings - UC-1.2.3
func (r *UserRepo) GetNotificationSettings(ctx context.Context, userID int64) (*domain.NotificationSettings, error) {
	var settings domain.NotificationSettings
	var channelsJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT user_id, new_game, new_build, new_post, channels
		FROM notification_settings WHERE user_id = $1`, userID).Scan(
		&settings.UserID, &settings.NewGame, &settings.NewBuild, &settings.NewPost, &channelsJSON)

	if err != nil {
		// Возвращаем дефолтные настройки
		defaultSettings := domain.DefaultNotificationSettings(userID)
		return &defaultSettings, nil
	}

	if err := json.Unmarshal(channelsJSON, &settings.Channels); err != nil {
		settings.Channels = domain.DefaultChannels()
	}
	return &settings, nil
}

// CreateResetToken - UC-1.1.3 (TTL 1 час)
func (r *UserRepo) CreateResetToken(ctx context.Context, userID int64, token string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO password_reset_tokens (token, user_id, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '1 hour')
		ON CONFLICT (user_id) DO UPDATE SET token = $1, expires_at = NOW() + INTERVAL '1 hour'`,
		token, userID)
	return err
}

// ResetPassword - UC-1.1.3 сброс пароля и всех сессий
func (r *UserRepo) ResetPassword(ctx context.Context, token, newHash string) (bool, int64, error) {
	var userID int64
	err := r.db.QueryRow(ctx, `
		UPDATE users SET hash = $2
		WHERE id = (
			SELECT user_id FROM password_reset_tokens 
			WHERE token = $1 AND expires_at > NOW()
		)
		RETURNING id`, token, newHash).Scan(&userID)

	if err != nil {
		return false, 0, err
	}

	// Удаляем использованный токен
	r.db.Exec(ctx, "DELETE FROM password_reset_tokens WHERE token = $1", token)
	return true, userID, nil
}

// UserSubmission - сабмит пользователя для UC-1.2.2
type UserSubmission struct {
	ID          int64     `json:"id"`
	JamTitle    string    `json:"jam_title"`
	JamSlug     string    `json:"jam_slug"`
	GameTitle   string    `json:"game_title"`
	GameSlug    string    `json:"game_slug"`
	Status      string    `json:"status"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// GetUserSubmissions - UC-1.2.2 список сабмитов пользователя
func (r *UserRepo) GetUserSubmissions(ctx context.Context, userID int64, page, limit int) ([]UserSubmission, int, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	// Пока используем заглушку, в будущем будет JOIN с реальными таблицами jams и games
	rows, err := r.db.Query(ctx, `
		SELECT 
			id,
			COALESCE(jam_title, 'Test Jam') as jam_title,
			COALESCE(jam_slug, 'test-jam') as jam_slug,
			COALESCE(game_title, 'My Game') as game_title,
			COALESCE(game_slug, 'my-game') as game_slug,
			status,
			submitted_at
		FROM user_submissions 
		WHERE user_id = $1
		ORDER BY submitted_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Логируем ошибку закрытия, но не возвращаем её
		}
	}()

	var submissions []UserSubmission
	for rows.Next() {
		var s UserSubmission
		if err := rows.Scan(&s.ID, &s.JamTitle, &s.JamSlug, &s.GameTitle, &s.GameSlug, &s.Status, &s.SubmittedAt); err != nil {
			return nil, 0, err
		}
		submissions = append(submissions, s)
	}

	// Подсчет общего количества
	var total int
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM user_submissions WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return submissions, total, nil
}

// Ping - проверка соединения с БД
func (r *UserRepo) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// GetPoolStats - получение статистики connection pool
func (r *UserRepo) GetPoolStats() map[string]interface{} {
	stats := r.db.Stat()
	return map[string]interface{}{
		"acquired_conns":      stats.AcquiredConns(),
		"constructing_conns": stats.ConstructingConns(),
		"idle_conns":          stats.IdleConns(),
		"max_conns":           stats.MaxConns(),
		"total_conns":         stats.TotalConns(),
		"acquire_count":       stats.AcquireCount(),
		"acquire_duration":    stats.AcquireDuration().String(),
		"canceled_acquire_count": stats.CanceledAcquireCount(),
		"empty_acquire_count":    stats.EmptyAcquireCount(),
	}
}

// UpdateAvatar - обновление аватара пользователя
func (r *UserRepo) UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET avatar = $2 WHERE id = $1`, userID, avatarURL)
	return err
}
