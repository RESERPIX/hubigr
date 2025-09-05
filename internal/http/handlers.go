package http

import (
	"mime/multipart"
	"strings"

	"github.com/RESERPIX/hubigr/internal/captcha"
	"github.com/RESERPIX/hubigr/internal/domain"
	"github.com/RESERPIX/hubigr/internal/logger"
	"github.com/RESERPIX/hubigr/internal/metrics"
	"github.com/RESERPIX/hubigr/internal/ratelimit"
	"github.com/RESERPIX/hubigr/internal/security"
	"github.com/RESERPIX/hubigr/internal/store"
	"github.com/RESERPIX/hubigr/internal/utils"
	"github.com/RESERPIX/hubigr/internal/validation"
	"github.com/gofiber/fiber/v2"
)



type Handlers struct {
	userRepo       *store.UserRepo
	refreshRepo    *store.RefreshTokenRepo
	limiter        *ratelimit.RedisLimiter
	emailSender    EmailSender
	avatarUploader AvatarUploader
	jwtSecret      string
	turnstile      *captcha.TurnstileService
	// TTL Policies
	accessTokenTTL  int
	refreshTokenTTL int
}

type AvatarUploader interface {
	UploadAvatar(userID int64, file *multipart.FileHeader) (string, error)
	DeleteAvatar(avatarURL string) error
}

type EmailSender interface {
	SendVerificationEmail(to, token string) error
	SendPasswordResetEmail(to, token string) error
}

func NewHandlers(userRepo *store.UserRepo, refreshRepo *store.RefreshTokenRepo, limiter *ratelimit.RedisLimiter, emailSender EmailSender, avatarUploader AvatarUploader, jwtSecret string, turnstile *captcha.TurnstileService, accessTTL, refreshTTL int) *Handlers {
	return &Handlers{userRepo: userRepo, refreshRepo: refreshRepo, limiter: limiter, emailSender: emailSender, avatarUploader: avatarUploader, jwtSecret: jwtSecret, turnstile: turnstile, accessTokenTTL: accessTTL, refreshTokenTTL: refreshTTL}
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
		logger.Error("Failed to send verification email", "error", err, "email", utils.SanitizeEmail(req.Email))
	} else {
		metrics.IncrementEmailSent()
	}
	
	// Метрика регистрации
	metrics.IncrementUserRegistered()

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
		metrics.IncrementLoginAttempt(false)
		return c.Status(401).JSON(domain.NewError("unauthorized", "Неверный email или пароль"))
	}

	// Проверка пароля
	if !security.CheckPassword(user.Hash, req.Password) {
		metrics.IncrementLoginAttempt(false)
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

	// Генерация access token
	accessToken, err := security.SignJWT(user.ID, string(user.Role), user.Nick, h.jwtSecret, h.accessTokenTTL)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка создания токена"))
	}

	// Создание refresh token
	deviceInfo := c.Get("User-Agent")
	ipAddress := c.IP()
	refreshToken, err := h.refreshRepo.Create(c.Context(), user.ID, deviceInfo, ipAddress, h.refreshTokenTTL)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка создания refresh токена"))
	}

	// Очистка хеша пароля из ответа
	user.Hash = ""
	
	// Метрика успешного входа
	metrics.IncrementLoginAttempt(true)

	return c.JSON(domain.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
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

// Logout - UC-1.1.4 отзыв всех refresh токенов
func (h *Handlers) Logout(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if ok {
		h.refreshRepo.RevokeUserTokens(c.Context(), userID)
	}
	return c.JSON(fiber.Map{"message": "Вы успешно вышли из системы"})
}

// Health - проверка здоровья сервиса
func (h *Handlers) Health(c *fiber.Ctx) error {
	ctx := c.Context()
	status := "healthy"
	statusCode := 200
	
	// Проверка БД
	dbErr := h.userRepo.Ping(ctx)
	if dbErr != nil {
		status = "unhealthy"
		statusCode = 503
	}

	// Проверка Redis
	redisErr := h.limiter.Ping(ctx)
	if redisErr != nil {
		status = "unhealthy"
		statusCode = 503
	}

	// Проверка метрик
	m := metrics.GetMetrics()
	snapshot := m.GetSnapshot()
	
	// Простые алерты
	alerts := []string{}
	if failedLogins, ok := snapshot["failed_logins"].(int64); ok && failedLogins > 10 {
		alerts = append(alerts, "High failed login rate")
	}
	if len(alerts) > 0 {
		status = "degraded"
	}

	response := fiber.Map{
		"status":   status,
		"database": map[string]interface{}{"status": "up", "error": dbErr},
		"redis":    map[string]interface{}{"status": "up", "error": redisErr},
		"version":  "1.0.0",
		"alerts":   alerts,
		"metrics":  snapshot,
	}
	
	if dbErr != nil {
		response["database"] = map[string]interface{}{"status": "down", "error": dbErr.Error()}
	}
	if redisErr != nil {
		response["redis"] = map[string]interface{}{"status": "down", "error": redisErr.Error()}
	}

	return c.Status(statusCode).JSON(response)
}

// Metrics - эндпоинт для метрик
func (h *Handlers) Metrics(c *fiber.Ctx) error {
	m := metrics.GetMetrics()
	return c.JSON(m.GetSnapshot())
}

// PrometheusMetrics - эндпоинт для Prometheus
func (h *Handlers) PrometheusMetrics(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/plain; charset=utf-8")
	return c.SendString(metrics.PrometheusFormat())
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
		logger.Error("Failed to send verification email", "error", err, "email", utils.SanitizeEmail(req.Email))
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
		logger.Error("Failed to send reset email", "error", err, "email", utils.SanitizeEmail(req.Email))
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
	success, userID, err := h.userRepo.ResetPassword(c.Context(), req.Token, hash)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка сброса пароля"))
	}

	if !success {
		return c.Status(400).JSON(domain.NewError("invalid_token", "Токен недействителен или истек"))
	}

	// Отзываем все refresh токены при смене пароля
	h.refreshRepo.RevokeUserTokens(c.Context(), userID)

	return c.JSON(fiber.Map{"message": "Пароль успешно изменен"})
}

// GetMySubmissions - UC-1.2.2 список сабмитов пользователя
func (h *Handlers) GetMySubmissions(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка получения ID пользователя"))
	}

	// Пагинация с валидацией
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	
	// Валидация пагинации
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

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
		if err := h.avatarUploader.DeleteAvatar(*user.Avatar); err != nil {
			logger.Error("Failed to delete old avatar", "error", err, "user_id", userID)
		}
	}

	return c.JSON(fiber.Map{
		"avatar_url": avatarURL,
		"message":    "Аватар обновлен",
	})
}

// RefreshToken обновляет access token через refresh token
func (h *Handlers) RefreshToken(c *fiber.Ctx) error {
	var req domain.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(domain.NewError("bad_request", "Неверный формат данных"))
	}

	// Проверяем и обновляем refresh token
	oldToken, newRefreshToken, err := h.refreshRepo.ValidateAndRotate(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(domain.NewError("unauthorized", "Недействительный refresh token"))
	}

	// Получаем пользователя
	user, err := h.userRepo.GetByID(c.Context(), oldToken.UserID)
	if err != nil {
		return c.Status(404).JSON(domain.NewError("not_found", "Пользователь не найден"))
	}

	// Проверяем что пользователь не заблокирован
	if user.IsBanned {
		return c.Status(403).JSON(domain.NewError("forbidden", "Аккаунт заблокирован"))
	}

	// Генерируем новый access token
	accessToken, err := security.SignJWT(user.ID, string(user.Role), user.Nick, h.jwtSecret, h.accessTokenTTL)
	if err != nil {
		return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка создания токена"))
	}

	// Очищаем хеш пароля
	user.Hash = ""

	return c.JSON(domain.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	})
}
