package errors

import (
	"github.com/RESERPIX/hubigr/internal/domain"
	"github.com/RESERPIX/hubigr/internal/logger"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler - централизованная обработка ошибок
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Внутренняя ошибка сервера"

	// Обработка Fiber ошибок
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Логирование ошибки
	errorMsg := "unknown error"
	if err != nil {
		errorMsg = err.Error()
	}
	logger.Error("Request error",
		"error", errorMsg,
		"method", c.Method(),
		"path", c.Path(),
		"ip", c.IP(),
		"status", code,
	)

	// Возврат ошибки клиенту
	return c.Status(code).JSON(domain.NewError("internal_error", message))
}