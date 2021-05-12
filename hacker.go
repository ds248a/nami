package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// главная страница
func (t *Controller) mainPage(c *gin.Context) {
	c.HTML(http.StatusOK, "main_page.tmpl", gin.H{})
}

// данные хакера
// принимает сообщение вида: /json/hacker?name='Alan Kay'
func (t *Controller) hacker(c *gin.Context) {
	name := c.Query("name")
	err := validate.Var(name, "required,gte=3,lte=50,excludesall=!$@#?%")
	if err != nil {

		// -- пример ведения лога
		// lg.LogMsg("db err msg").Query("user: sairos").Save()
		// lg.LogErr(err).Query("user: tomo").Save()

		c.JSON(http.StatusOK, gin.H{"error": "name_not_valid"})
		return
	}

	h, err := t.mHack.hacker(context.Background(), name)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "empty_list"})
		return
	}

	c.JSON(http.StatusOK, h)
}

// список хакеров
// принимает сообщение вида: /json/hackers
func (t *Controller) hackersList(c *gin.Context) {
	hl, ok := t.mHack.hackersList(context.Background())

	if !ok {
		c.JSON(http.StatusOK, gin.H{"error": "empty_list"})
		return
	}

	c.JSON(http.StatusOK, hl)
}

// создает заданное кол-во записей в наборе хакеров
// принимает сообщение вида: /new?count=1000
func (t *Controller) hackerNew(c *gin.Context) {
	count, err := strconv.Atoi(string(c.Params.ByName("count")))
	if err != nil || count < 1 || count > 10000 {
		c.JSON(http.StatusOK, gin.H{"error": "count_not_valid"})
		return
	}

	if err := t.mHack.hackerNew(context.Background(), count); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "count_not_valid"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": true})
}

// восстановление списка хакеров
// принимает сообщение вида: /recover
func (t *Controller) hackerRecover(c *gin.Context) {
	if err := t.mHack.hackerRecover(context.Background()); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "hacker_recover"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": true})
}

func (t *Controller) notFound(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"title": "404"})
}
