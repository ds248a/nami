package main

import (
	"fmt"

	"github.com/ds248a/nami/app"
)

func main() {
	nami := app.NewNami()
	fmt.Println(nami.A)
}
