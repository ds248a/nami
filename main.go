package main

import (
	"time"

	"github.com/ds248a/air/app"
	"github.com/go-playground/validator/v10"
)

const (
	sTimer time.Duration = 5 * time.Minute
	mTimer time.Duration = 10 * time.Minute
	lTimer time.Duration = 20 * time.Minute
	xTimer time.Duration = 60 * time.Minute
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
	cNami.router(r)

	// запуск HTTP сервера
	app.StartHTTP(r)

	// обработка прерываний сервера HTTP
	app.Signal()
}
