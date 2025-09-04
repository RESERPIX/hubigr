package http

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/yourorg/users/internal/domain"
	"github.com/yourorg/users/internal/security"
	"github.com/yourorg/users/internal/store"
)

type Handlers struct {
	Users *store.UserRepo
}

func jsonErr(c *fiber.Ctx, code int, codeStr, msg string) error {
	return c.Status(code).JSON(domain.NewError(codeStr, msg))
}

// --- Auth ---

type signUpReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nick     string `json:"nick"`
}

func (h *Handlers) SignUp(c *fiber.Ctx) error {
	var in signUpReq
	if err := c.BodyParser(&in); err != nil {
		return jsonErr(c, 400, "bad_request", "invalid json")
	}
	if !validEmail(in.Email) {
		return jsonErr(c, 422, "validation_error", "invalid email")
	}
	if l := utf8.RuneCountInString(in.Password); l < 6 || l > 72 {
		return jsonErr(c, 422, "validation_error", "password length 6..72")
	}
	if strings.TrimSpace(in.Nick) == "" {
		return jsonErr(c, 422, "validation_error", "nick is required")
	}

	hash, err := security.HashPassword(in.Password)
	if err != nil {
		return jsonErr(c, 500, "hash_error", "internal error")
	}

	ctx := context.Background()
	id, err := h.Users.CreateUser(ctx, in.Email, hash, in.Nick)
	if err != nil {
		// предполагаем unique violation по email
		return jsonErr(c, 409, "conflict", "email already registered")
	}

	// создать verify token (1 час)
	token := randToken()
	_ = h.Users.UpsertVerifyToken(ctx, id, token, time.Hour)

	// TODO: отправка письма (заглушка)
	// sendEmail(in.Email, verifyURL(token))

	user := fiber.Map{
		"id": id, "email": strings.ToLower(in.Email), "nick": in.Nick,
		"role": "participant",
	}
	// ТЗ просит вернуть { user, token } у /signup — но логин до verify запрещён.
	// Решение: отдать token=empty и требовать верификацию.
	out := fiber.Map{"user": user, "token": ""}
	return c.JSON(out)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handlers) Login(c *fiber.Ctx) error {
	var in loginReq
	if err := c.BodyParser(&in); err != nil {
		return jsonErr(c, 400, "bad_request", "invalid json")
	}
	u, err := h.Users.GetByEmail(c.Context(), in.Email)
	if err != nil || u == nil {
		return jsonErr(c, 401, "unauthorized", "invalid email or password")
	}
	if !security.CheckPassword(u.PasswordHash, in.Password) {
		return jsonErr(c, 401, "unauthorized", "invalid email or password")
	}
	if u.IsBanned {
		return jsonErr(c, 403, "forbidden", "account is banned")
	}
	if !u.EmailVerified {
		return jsonErr(c, 401, "email_not_verified", "please verify your email")
	}

	tok, err := security.SignJWT(u.ID, u.Role, u.Nick)
	if err != nil {
		return jsonErr(c, 500, "token_error", "cannot sign token")
	}
	user := fiber.Map{
		"id": u.ID, "email": strings.ToLower(u.Email), "nick": u.Nick,
		"role": u.Role, "avatarUrl": u.AvatarURL, "bio": u.Bio, "links": u.Links,
	}
	return c.JSON(fiber.Map{"user": user, "token": tok})
}

func (h *Handlers) Logout(c *fiber.Ctx) error {
	// Стейтлесс JWT: на стороне сервера нечего очищать.
	return c.JSON(fiber.Map{"ok": true})
}

type verifyReq struct {
	Token string `json:"token"`
}

func (h *Handlers) VerifyEmail(c *fiber.Ctx) error {
	var in verifyReq
	if err := c.BodyParser(&in); err != nil || strings.TrimSpace(in.Token) == "" {
		return jsonErr(c, 400, "bad_request", "token required")
	}
	ok, err := h.Users.VerifyEmail(c.Context(), in.Token)
	if err != nil {
		return jsonErr(c, 500, "verify_error", "internal error")
	}
	if !ok {
		return jsonErr(c, 400, "invalid_token", "token invalid or expired")
	}
	return c.JSON(fiber.Map{"ok": true})
}

// заглушки сброса пароля (UC-1.1.3: TTL 1 час)
func (h *Handlers) ResetRequest(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) }
func (h *Handlers) ResetConfirm(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) }

// --- Profiles ---

func (h *Handlers) Me(c *fiber.Ctx) error {
	// В Шаге 2 добавим JWT-middleware и извлечём userID из токена.
	return jsonErr(c, 401, "unauthorized", "missing token")
}

type updateProfileReq struct {
	Nick      string   `json:"nick"`
	AvatarURL *string  `json:"avatarUrl"`
	Bio       *string  `json:"bio"`
	Links     []string `json:"links"`
}

func (h *Handlers) UpdateProfile(c *fiber.Ctx) error {
	// Здесь тоже потребуется userID из JWT — добавим в Шаге 2.
	return jsonErr(c, 401, "unauthorized", "missing token")
}

func (h *Handlers) MySubmissions(c *fiber.Ctx) error {
	// Понадобится userID из JWT — добавим в Шаге 2.
	return jsonErr(c, 401, "unauthorized", "missing token")
}

// --- helpers ---

func validEmail(s string) bool {
	s = strings.TrimSpace(s)
	return strings.Count(s, "@") == 1 && len(s) >= 6 && strings.Contains(s, ".")
}

func randToken() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 48)
	for i := range b {
		b[i] = alphabet[int(time.Now().UnixNano()+int64(i))%len(alphabet)]
	}
	return string(b)
}
