package blog

import (
	"fmt"

	"github.com/ds248a/nami/app"
)

type ModelBlog struct {
}

func InitPost() {

}

func Post() {
	app.NamiPlus()
	fmt.Println("post:", app.NamiA())
}
