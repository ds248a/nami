package main

import (
	"fmt"

	"github.com/ds248a/nami/app"
	"github.com/ds248a/nami/blog"
)

func main() {
	nami := app.NewNami()
	fmt.Println(nami.A)

	blog.Post()

	app.NamiPlus()
	fmt.Println("main:", app.NamiA())

	blog.Post()
}
