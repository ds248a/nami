package db

import (
	"fmt"

	"github.com/ds248a/nami/app"
)

func InitRedis() {

}

func Redis() {
	app.NamiPlus()
	fmt.Println("redis:", app.NamiA())
}
