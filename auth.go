package main

import fiber "github.com/gofiber/fiber/v2"

func loadAuth(app *fiber.App) {
	auth := app.Group("/auth")
	auth.Get("/config")
}
