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
	r.POST("/j/hackers", t.hackersList)
	r.POST("/j/hacker", t.hacker)
	r.POST("/j/new", t.hackerNew)
	r.POST("/j/recover", t.hackerRecover)

	//r.NoMethod(t.notFound)
	//r.NoRoute(t.notFound)

	// ***********************************
	// ***  Подгововка списка хакеров  ***
	// t.mHack.hackerRecover(context.Background())
}
