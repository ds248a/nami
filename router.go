package main

import (
	"github.com/gin-gonic/gin"
)

// --------------------------------
//    Router
// --------------------------------

type Controller struct {
	mHack *HackerModel
}

func (t *Controller) router(r *gin.Engine) {
	r.GET("/", t.mainPage)
	r.GET("/json/hackers", t.hackersList)
	r.GET("/json/hacker", t.hacker)
	r.GET("/new", t.hackerNew)
	r.GET("/recover", t.hackerRecover)

	r.NoMethod(t.notFound)
	r.NoRoute(t.notFound)

	// ***********************************
	// ***  Подгововка списка хакеров  ***
	// t.mHack.hackerRecover(context.Background())
}
