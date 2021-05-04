package blog

import (
	"fmt"

	"github.com/ds248a/nami/app"
)

func InitPost() {

}

func Post() {
	fmt.Println("post:", app.NamiA())
	app.NamiPlus()
	fmt.Println("post:", app.NamiA())
}
