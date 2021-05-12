package app

import (
	"strconv"
)

// --------------------------------
//    Формат вывода
// --------------------------------

type H map[string]interface{}

// отправка данных в формате JSON
/*
func JFWrite(code int, ctx *fasthttp.RequestCtx, obj interface{}) {
	ctx.SetContentType("application/json;charset=utf-8")
	ctx.SetStatusCode(code)
	if err := json.NewEncoder(ctx).Encode(&obj); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}
*/

// отправка ошибки в формате JSON
// данный формат используется в библиотеке language фреймворка VueJS
// организует могоязыковую поддержку отображения ошибки на стороне клиента
/*
func JFError(code int, ctx *fasthttp.RequestCtx, msg string) {
	ctx.SetContentType("application/json;charset=utf-8")
	ctx.SetStatusCode(code)
	if err := json.NewEncoder(ctx).Encode(&H{"error": msg}); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}
*/

func I16S(i uint16) string {
	return strconv.FormatUint(uint64(i), 10)
}

func I3S(i uint32) string {
	return strconv.FormatUint(uint64(i), 10)
}

func I6S(i uint64) string {
	return strconv.FormatUint(i, 10)
}
