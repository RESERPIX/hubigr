package store

import (
	"context"
	"time"

	"github.com/RESERPIX/hubigr/internal/domain"
	"github.com/RESERPIX/hubigr/internal/security"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepo struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepo(db *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, userID int64, deviceInfo, ipAddress string, ttlDays int) (string, error) {
	token, err := security.GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	hash, err := security.HashRefreshToken(token)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Duration(ttlDays) * 24 * time.Hour)

	_, err = r.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, device_info, ip_address)
		VALUES ($1, $2, $3, $4, $5)`,
		userID, hash, expiresAt, deviceInfo, ipAddress)
	
	return token, err
}

func (r *RefreshTokenRepo) ValidateAndRotate(ctx context.Context, token string) (*domain.RefreshToken, string, error) {
	// Начинаем транзакцию для атомарности
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
			// Логируем ошибку rollback, но не возвращаем её
		}
	}()

	// Хешируем токен для поиска по индексу
	tokenHash, err := security.HashRefreshToken(token)
	if err != nil {
		return nil, "", err
	}

	// Прямой поиск по хешу - O(1) вместо O(n)
	var rt domain.RefreshToken
	err = tx.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, created_at, device_info, ip_address
		FROM refresh_tokens 
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
		FOR UPDATE`, tokenHash).Scan(
		&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, 
		&rt.CreatedAt, &rt.DeviceInfo, &rt.IPAddress)
	
	if err != nil {
		return nil, "", err
	}

	// Проверяем токен
	if !security.VerifyRefreshToken(rt.TokenHash, token) {
		return nil, "", pgx.ErrNoRows
	}

	// Отзываем старый токен
	_, err = tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1`, rt.ID)
	if err != nil {
		return nil, "", err
	}

	// Создаем новый токен
	newToken, err := r.createInTx(ctx, tx, rt.UserID, *rt.DeviceInfo, *rt.IPAddress, 7)
	if err != nil {
		return nil, "", err
	}

	// Подтверждаем транзакцию
	if err = tx.Commit(ctx); err != nil {
		return nil, "", err
	}

	return &rt, newToken, nil
}

// createInTx создает токен в рамках транзакции
func (r *RefreshTokenRepo) createInTx(ctx context.Context, tx pgx.Tx, userID int64, deviceInfo, ipAddress string, ttlDays int) (string, error) {
	token, err := security.GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	hash, err := security.HashRefreshToken(token)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Duration(ttlDays) * 24 * time.Hour)

	_, err = tx.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, device_info, ip_address)
		VALUES ($1, $2, $3, $4, $5)`,
		userID, hash, expiresAt, deviceInfo, ipAddress)
	
	return token, err
}

func (r *RefreshTokenRepo) RevokeUserTokens(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}