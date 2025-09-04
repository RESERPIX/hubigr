package http

import (
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, h *Handlers) {
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Get("/readyz", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ready"}) })

	// Auth
	auth := app.Group("")
	auth.Post("/signup", h.SignUp)
	// rate-limit: 5 попыток/мин (из ТЗ)
	auth.Post("/login", limiter.New(limiter.Config{Max: 5, Expiration: 60 * 1e9}), h.Login)
	auth.Post("/logout", h.Logout)
	auth.Post("/verify-email", h.VerifyEmail)            // { token }
	auth.Post("/reset-password/request", h.ResetRequest) // заглушка
	auth.Post("/reset-password/confirm", h.ResetConfirm)

	// Users
	app.Get("/me", h.Me) // авторизация по JWT (добавим в шаге 2)
	app.Put("/profile", h.UpdateProfile)
	app.Get("/profile/submissions", h.MySubmissions)
}
