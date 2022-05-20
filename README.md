# nami

Пример web сервера в MVC стиле для более комфортного перехода с PHP на Go.  
Включает основные необходимые компоненты:

go-cache - локальное шранилище типа ключ/значение  
redis/v8 - подержка распределенного хранилища
pgx/v4 - драйвер СУБД PostgreSQL  
validator/v10 - валидатор HTTP запросов
glg - лог пакет по умолчанию  


```go
package main

import (
	"github.com/ds248a/nami/app"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func main() {
	defer app.Close()

	cNami := Controller{
		mHack: &HackerModel{
			db: app.Redis(),
			lc: app.Cache(),
		},
	}

	r := app.Router()
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")
	cNami.router(r)

	// запуск HTTP сервера
	app.StartHTTP(r)

	// обработка прерываний сервера HTTP
	app.Signal()
}
```