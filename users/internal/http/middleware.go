package http

import (
	"github.com/gofiber/fiber/v2"
)

func Common(app *fiber.App) {
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // сузим потом под фронт
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
}
