package http

import (
	"mime/multipart"
	"strings"

	"github.com/RESERPIX/hubigr/internal/captcha"
	"github.com/RESERPIX/hubigr/internal/domain"
	"github.com/RESERPIX/hubigr/internal/logger"
	"github.com/RESERPIX/hubigr/internal/ratelimit"
	"github.com/RESERPIX/hubigr/internal/security"
	"github.com/RESERPIX/hubigr/internal/store"
	"github.com/RESERPIX/hubigr/internal/validation"
	"github.com/gofiber/fiber/v2"
)

// maskEmail маскирует email для безопасного логирования
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}
	local := parts[0]
	domain := parts[1]
	
	if len(local) <= 2 {
		return "***@" + domain
	}
	return local[:2] + "***@" + domain
}

type Handlers struct {
	userRepo       *store.UserRepo
	limiter        *ratelimit.RedisLimiter
	emailSender    EmailSender
	avatarUploader AvatarUploader
	jwtSecret      string
	turnstile      *captcha.TurnstileService
}

type AvatarUploader interface {
	UploadAvatar(userID int64, file *multipart.FileHeader) (string, error)
	DeleteAvatar(avatarURL string) error
}

type EmailSender interface {
	SendVerificationEmail(to, token string) error
	SendPasswordResetEmail(to, token string) error
}

func NewHandlers(userRepo *store.UserRepo, limiter *ratelimit.RedisLimiter, emailSender EmailSender, avatarUploader AvatarUploader, jwtSecret string, turnstile *captcha.TurnstileService) *Handlers {
	return &Handlers{userRepo: userRepo, limiter: limiter, emailSender: emailSender, avatarUploader: avatarUploader, jwtSecret: jwtSecret, turnstile: turnstile}
}

// SignUp - UC-1.1.1 из ТЗ
func (h *Handlers) SignUp(c *fiber.Ctx) error {
	var req domain.SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(domain.NewError("bad_request", "Неверный формат данных"))
	}

	// Валидация согласно ТЗ
	if errors := validation.ValidateSignUp(req); len(errors) > 0 {
		return c.Status(422).JSON(domain.NewError("validation_error", strings.Join(errors, "; ")))
	}

	// Хеширование пароля
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка сервера"))
	}

	// Создание пользователя
	userID, err := h.userRepo.CreateUser(c.Context(), req, hash)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return c.Status(409).JSON(domain.NewError("conflict", "Пользователь с таким email уже зарегистрирован"))
		}
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка создания пользователя"))
	}

	// Создание токена подтверждения (TTL 1 час)
	token, err := security.GenerateToken()
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка генерации токена"))
	}
	if err := h.userRepo.CreateVerifyToken(c.Context(), userID, token); err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка создания токена"))
	}

	// Отправка email с подтверждением
	if err := h.emailSender.SendVerificationEmail(req.Email, token); err != nil {
		logger.Error("Failed to send verification email", "error", err, "email", maskEmail(req.Email))
	}

	return c.JSON(fiber.Map{
		"message": "Мы отправили ссылку для подтверждения на email",
		"user_id": userID,
	})
}

// Login - UC-1.1.2 из ТЗ
func (h *Handlers) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(domain.NewError("bad_request", "Неверный формат данных"))
	}

	// Получение пользователя
	user, err := h.userRepo.GetByEmail(c.Context(), req.Email)
	if err != nil || user == nil {
		return c.Status(401).JSON(domain.NewError("unauthorized", "Неверный email или пароль"))
	}

	// Проверка пароля
	if !security.CheckPassword(user.Hash, req.Password) {
		return c.Status(401).JSON(domain.NewError("unauthorized", "Неверный email или пароль"))
	}

	// Проверка бана
	if user.IsBanned {
		return c.Status(403).JSON(domain.NewError("forbidden", "Аккаунт заблокирован"))
	}

	// Проверка подтверждения email
	if !user.EmailVerified {
		return c.Status(401).JSON(domain.NewError("email_not_verified", "Подтвердите email для входа"))
	}

	// Генерация JWT
	token, err := security.SignJWT(user.ID, string(user.Role), user.Nick, h.jwtSecret)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка создания токена"))
	}

	// Очистка хеша пароля из ответа
	user.Hash = ""

	return c.JSON(domain.AuthResponse{
		User:  *user,
		Token: token,
	})
}

// VerifyEmail - UC-1.1.1 из ТЗ
func (h *Handlers) VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(400).JSON(domain.NewError("bad_request", "Токен обязателен"))
	}

	success, err := h.userRepo.VerifyEmail(c.Context(), token)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка подтверждения"))
	}

	if !success {
		return c.Status(400).JSON(domain.NewError("invalid_token", "Токен недействителен или истек"))
	}

	return c.JSON(fiber.Map{"message": "Email успешно подтвержден"})
}

// Logout - UC-1.1.4 из ТЗ (stateless JWT)
func (h *Handlers) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Вы успешно вышли из системы"})
}

// Health - проверка здоровья сервиса
func (h *Handlers) Health(c *fiber.Ctx) error {
	// Проверка БД
	if err := h.userRepo.Ping(c.Context()); err != nil {
		return c.Status(503).JSON(fiber.Map{
			"status": "unhealthy",
			"database": "down",
			"error": err.Error(),
		})
	}

	// Проверка Redis
	if err := h.limiter.Ping(c.Context()); err != nil {
		return c.Status(503).JSON(fiber.Map{
			"status": "unhealthy",
			"redis": "down",
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "healthy",
		"database": "up",
		"redis": "up",
		"version": "1.0.0",
	})
}

// GetProfile - UC-1.2.2 из ТЗ
func (h *Handlers) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка получения ID пользователя"))
	}

	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(domain.NewError("not_found", "Пользователь не найден"))
	}

	// Очистка хеша пароля
	user.Hash = ""

	return c.JSON(user)
}

// UpdateProfile - UC-1.2.1 из ТЗ
func (h *Handlers) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка получения ID пользователя"))
	}

	var req domain.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(domain.NewError("bad_request", "Неверный формат данных"))
	}

	// Валидация согласно ТЗ
	if errors := validation.ValidateProfile(req); len(errors) > 0 {
		return c.Status(422).JSON(domain.NewError("validation_error", strings.Join(errors, "; ")))
	}

	if err := h.userRepo.UpdateProfile(c.Context(), userID, req); err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка обновления профиля"))
	}

	return c.JSON(fiber.Map{"message": "Профиль обновлен"})
}

// UpdateNotifications - UC-1.2.3 из ТЗ
func (h *Handlers) UpdateNotifications(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка получения ID пользователя"))
	}

	var settings domain.NotificationSettings
	if err := c.BodyParser(&settings); err != nil {
		return c.Status(400).JSON(domain.NewError("bad_request", "Неверный формат данных"))
	}

	settings.UserID = userID
	if err := h.userRepo.UpdateNotificationSettings(c.Context(), userID, settings); err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка сохранения настроек"))
	}

	return c.JSON(fiber.Map{"message": "Настройки обновлены"})
}

// GetNotifications - получение настроек уведомлений
func (h *Handlers) GetNotifications(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка получения ID пользователя"))
	}

	settings, err := h.userRepo.GetNotificationSettings(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка получения настроек"))
	}

	return c.JSON(settings)
}

// ResendVerification - повторная отправка письма подтверждения
func (h *Handlers) ResendVerification(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(domain.NewError("bad_request", "Неверный формат данных"))
	}

	// Получение пользователя
	user, err := h.userRepo.GetByEmail(c.Context(), req.Email)
	if err != nil || user == nil {
		// Не раскрываем существование email
		return c.JSON(fiber.Map{"message": "Если email существует, мы отправим новую ссылку"})
	}

	// Проверяем, нуждается ли в подтверждении
	if user.EmailVerified {
		return c.Status(400).JSON(domain.NewError("already_verified", "Email уже подтвержден"))
	}

	// Создание нового токена
	token, err := security.GenerateToken()
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка генерации токена"))
	}
	if err := h.userRepo.CreateVerifyToken(c.Context(), user.ID, token); err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка создания токена"))
	}

	// Отправка email
	if err := h.emailSender.SendVerificationEmail(req.Email, token); err != nil {
		logger.Error("Failed to send verification email", "error", err, "email", maskEmail(req.Email))
	}

	return c.JSON(fiber.Map{"message": "Новая ссылка отправлена на email"})
}

// ResetPasswordRequest - UC-1.1.3 запрос сброса пароля
func (h *Handlers) ResetPasswordRequest(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(domain.NewError("bad_request", "Неверный формат данных"))
	}

	// Получение пользователя
	user, err := h.userRepo.GetByEmail(c.Context(), req.Email)
	if err != nil || user == nil {
		// Не раскрываем существование email для безопасности
		return c.JSON(fiber.Map{"message": "Мы отправили ссылку на email"})
	}

	// Создание токена сброса (TTL 1 час)
	token, err := security.GenerateToken()
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка генерации токена"))
	}
	if err := h.userRepo.CreateResetToken(c.Context(), user.ID, token); err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка создания токена"))
	}

	// Отправка email
	if err := h.emailSender.SendPasswordResetEmail(req.Email, token); err != nil {
		logger.Error("Failed to send reset email", "error", err, "email", maskEmail(req.Email))
	}

	return c.JSON(fiber.Map{"message": "Мы отправили ссылку на email"})
}

// ResetPasswordConfirm - UC-1.1.3 подтверждение сброса пароля
func (h *Handlers) ResetPasswordConfirm(c *fiber.Ctx) error {
	var req struct {
		Token           string `json:"token"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(domain.NewError("bad_request", "Неверный формат данных"))
	}

	// Валидация пароля
	if req.Password != req.ConfirmPassword {
		return c.Status(422).JSON(domain.NewError("validation_error", "Пароли должны совпадать"))
	}
	if l := len(req.Password); l < 6 || l > 20 {
		return c.Status(422).JSON(domain.NewError("validation_error", "Пароль должен содержать от 6 до 20 символов"))
	}

	// Хеширование нового пароля
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка сервера"))
	}

	// Сброс пароля и всех сессий
	success, err := h.userRepo.ResetPassword(c.Context(), req.Token, hash)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка сброса пароля"))
	}

	if !success {
		return c.Status(400).JSON(domain.NewError("invalid_token", "Токен недействителен или истек"))
	}

	return c.JSON(fiber.Map{"message": "Пароль успешно изменен"})
}

// GetMySubmissions - UC-1.2.2 список сабмитов пользователя
func (h *Handlers) GetMySubmissions(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка получения ID пользователя"))
	}

	// Пагинация
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	submissions, total, err := h.userRepo.GetUserSubmissions(c.Context(), userID, page, limit)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка получения сабмитов"))
	}

	return c.JSON(fiber.Map{
		"submissions": submissions,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// UploadAvatar - UC-1.2.1 загрузка аватара
func (h *Handlers) UploadAvatar(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка получения ID пользователя"))
	}

	// Получение файла из формы
	file, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(400).JSON(domain.NewError("bad_request", "Файл не найден"))
	}

	// Получение текущего аватара для удаления
	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(domain.NewError("not_found", "Пользователь не найден"))
	}

	// Загрузка нового аватара
	avatarURL, err := h.avatarUploader.UploadAvatar(userID, file)
	if err != nil {
		return c.Status(422).JSON(domain.NewError("upload_error", err.Error()))
	}

	// Обновление профиля
	if err := h.userRepo.UpdateAvatar(c.Context(), userID, avatarURL); err != nil {
		// Удаляем загруженный файл при ошибке
		h.avatarUploader.DeleteAvatar(avatarURL)
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка обновления профиля"))
	}

	// Удаляем старый аватар
	if user.Avatar != nil && *user.Avatar != "" {
		h.avatarUploader.DeleteAvatar(*user.Avatar)
	}

	return c.JSON(fiber.Map{
		"avatar_url": avatarURL,
		"message":    "Аватар обновлен",
	})
}
