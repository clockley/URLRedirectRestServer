package main

import (
	dbhelper "dwarfRestServer/db"
	"dwarfRestServer/urlcheck"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"os"
	"os/signal"
	"syscall"
)

func initRestAPI(db *dbhelper.DatabaseConnection) {
	app := fiber.New()
	app.Use(cors.New())

	app.Get("/:url", func(c *fiber.Ctx) error {
		hi := db.GetHashInfo(c.Params("url"))
		if hi == nil {
			return c.Status(fiber.StatusNotFound).SendString("Unable to find or load url")
		}
		return c.Redirect(hi.TargetUrl, 302)
	})

	app.Post("/shorten", func(c *fiber.Ctx) error {
		var payload dbhelper.Payload
		if err := defaults.Set(&payload); err != nil {
			panic(err)
		}
		err := c.BodyParser(&payload)

		if err != nil {
			return c.SendStatus(422)
		}

		if !urlcheck.IsSafeURL(payload.Url) {
			return c.SendStatus(422)
		}

		rowInfo := db.LookupURL(payload.Url)

		if rowInfo.TargetUrl == "" {
			db.CreateShortUrl(&payload)
		} else {
			payload.ShortUrl = fmt.Sprintf("https://%s/%s", rowInfo.Domain, rowInfo.HashId)
			return c.JSON(payload)
		}

		return c.JSON(payload)
	})

	app.Listen(":3000")
}

func main() {
	dbcon := new(dbhelper.DatabaseConnection).ConnectToDb()
	sigCh := make(chan os.Signal)
	go func() {
		select {
		case <-sigCh:
			dbcon.Close()
			os.Exit(0)
		}
	}()
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	initRestAPI(dbcon)
}
