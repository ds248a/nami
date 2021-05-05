package main

import (
	"fmt"

	"github.com/ds248a/nami/app"
	"github.com/ds248a/nami/blog"
	"github.com/ds248a/nami/db"
)

func main() {
	app.NewNami()

	blog.Post()

	app.NamiPlus()
	fmt.Println("main:", app.NamiA())

	app.NewNami()

	db.Redis()
	blog.Post()
}
