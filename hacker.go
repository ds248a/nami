package main

import (
	"context"
	"fmt"
	"net/http"
	// "strconv"

	"github.com/gin-gonic/gin"
)

type jHacker struct {
	Count int    `form:"count" json:"count" xml:"count"` //   binding:"required"
	Name  string `form:"name"  json:"name"  xml:"name"`
}

// главная страница
func (t *Controller) mainPage(c *gin.Context) {
	fmt.Println("mainPage")
	c.HTML(http.StatusOK, "main_page.tmpl", gin.H{})
}

// обработка параметров запроса
func hackerParse(c *gin.Context) (*jHacker, error) {
	js := &jHacker{}
	if err := c.ShouldBindJSON(js); err != nil {
		return nil, err
	}

	if len(js.Name) > 0 {
		if err := validate.Var(js.Name, "required,gte=3,lte=50,excludesall=!$@#?%"); err != nil {
			return nil, err
		}
	}

	fmt.Printf("%+v\n", js)

	return js, nil
}

// список хакеров
// POST /j/hackers
func (t *Controller) hackersList(c *gin.Context) {
	hl, ok := t.mHack.hackersList(context.Background())
	if !ok {
		c.JSON(http.StatusOK, gin.H{"error": "empty_list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hl})
}

// данные хакера
// POST /j/hacker name='Alan Kay'
func (t *Controller) hacker(c *gin.Context) {
	js, err := hackerParse(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "name_not_valid"})
		return
	}

	// -- пример ведения лога
	// lg.LogMsg("db err msg").Query("user: sairos").Save()
	// lg.LogErr(err).Query("user: tomo").Save()

	h, err := t.mHack.hacker(context.Background(), js.Name)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "empty_list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": h})
}

// создает заданное кол-во записей в наборе хакеров
// POST /j/new count=5
func (t *Controller) hackerNew(c *gin.Context) {
	js, err := hackerParse(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "count_not_valid"})
		return
	}

	if err := t.mHack.hackerNew(context.Background(), js.Count); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "count_not_valid"})
		return
	}

	hl, ok := t.mHack.hackersList(context.Background())
	if !ok {
		c.JSON(http.StatusOK, gin.H{"error": "empty_list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hl})
}

// восстановление списка хакеров
// POST /j/recover
func (t *Controller) hackerRecover(c *gin.Context) {
	hl, err := t.mHack.hackerRecover(context.Background())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "hacker_recover"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hl})
}

func (t *Controller) notFound(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"title": "404"})
}
