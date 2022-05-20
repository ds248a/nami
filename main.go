package main

import (
	"github.com/ds248a/nami/app"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func main() {
	defer app.Close()

	cNami := Controller{
		mHack: &HackerModel{
			db: app.Redis(),
			lc: app.Cache(),
		},
	}

	r := app.Router()
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")
	cNami.router(r)

	// запуск HTTP сервера
	app.StartHTTP(r)

	// обработка прерываний сервера HTTP
	app.Signal()
}
