-- Расширения PostgreSQL
CREATE EXTENSION IF NOT EXISTS citext;

-- Пользователи (DM-3.1 из ТЗ)
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email CITEXT UNIQUE NOT NULL,
    hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'participant' CHECK (role IN ('participant', 'jury', 'moderator', 'admin', 'organizer')),
    nick TEXT NOT NULL CHECK (char_length(nick) >= 2 AND char_length(nick) <= 50),
    avatar TEXT,
    bio TEXT CHECK (char_length(bio) <= 200),
    links JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_banned BOOLEAN NOT NULL DEFAULT false,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    privacy_settings JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

-- Токены подтверждения email (UC-1.1.1 - TTL 1 час)
CREATE TABLE IF NOT EXISTS email_verify_tokens (
    token TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_verify_tokens_user ON email_verify_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_verify_tokens_expires ON email_verify_tokens (expires_at);

-- Токены сброса пароля (UC-1.1.3 - TTL 1 час)
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    token TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_reset_tokens_user ON password_reset_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_reset_tokens_expires ON password_reset_tokens (expires_at);

-- Настройки уведомлений (DM-3.15 из ТЗ)
CREATE TABLE IF NOT EXISTS notification_settings (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    new_game BOOLEAN NOT NULL DEFAULT true,
    new_build BOOLEAN NOT NULL DEFAULT true,
    new_post BOOLEAN NOT NULL DEFAULT true,
    channels JSONB NOT NULL DEFAULT '{"in_app": true}'::jsonb
);

-- Подписки на разработчиков (DM-3.14 из ТЗ)
CREATE TABLE IF NOT EXISTS follows (
    follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followed_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followed_id),
    CHECK (follower_id != followed_id)
);

CREATE INDEX IF NOT EXISTS idx_follows_followed ON follows (followed_id);

-- Сабмиты пользователей (заглушка для UC-1.2.2)
CREATE TABLE IF NOT EXISTS user_submissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    jam_title TEXT,
    jam_slug TEXT,
    game_title TEXT,
    game_slug TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'hidden')),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_submissions_user ON user_submissions (user_id, submitted_at DESC);

-- Тестовые данные для демо
INSERT INTO user_submissions (user_id, jam_title, jam_slug, game_title, game_slug, status, submitted_at) 
VALUES 
    (1, 'Spring Game Jam 2024', 'spring-jam-2024', 'My Awesome Game', 'my-awesome-game', 'active', NOW() - INTERVAL '2 days'),
    (1, 'Ludum Dare 54', 'ludum-dare-54', 'Pixel Adventure', 'pixel-adventure', 'active', NOW() - INTERVAL '1 month')
ON CONFLICT DO NOTHING;

-- Очистка истекших токенов
CREATE OR REPLACE FUNCTION cleanup_expired_tokens() RETURNS void AS $$
BEGIN
    DELETE FROM email_verify_tokens WHERE expires_at < NOW();
    DELETE FROM password_reset_tokens WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- Автоматическая очистка каждый час
-- SELECT cron.schedule('cleanup-tokens', '0 * * * *', 'SELECT cleanup_expired_tokens();');