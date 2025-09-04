package main

import (
	"context"

	"github.com/gofiber/fiber/v2"
	h "github.com/yourorg/users/internal/http"
	"github.com/yourorg/users/internal/store"
)

func main() {
	ctx := context.Background()
	db := store.MustOpen(ctx)
	repo := store.NewUserRepo(db)

	app := fiber.New(fiber.Config{
		EnablePrintRoutes: false,
		CaseSensitive:     true,
		AppName:           "users",
	})
	h.Common(app)
	handlers := &h.Handlers{Users: repo}
	h.Register(app, handlers)

	if err := app.Listen(":8080"); err != nil {
		panic(err)
	}
}
