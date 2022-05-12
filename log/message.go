package log

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

// --------------------------------
//    Message Log
// --------------------------------

type Message struct {
	File string `db:"file" json:"file"`
	Line int    `db:"line" json:"line"`
	Fnct string `db:"function" json:"fnct"`
	Msg  []byte `db:"message" json:"msg"`
	Qry  []byte `db:"query" json:"query"`
	Date int64  `db:"date" json:"date"`
}

// Формирование сообщения лога, для последующего сохранения
func newMessage(msg string) *Message {
	m := &Message{
		Msg:  []byte(msg),
		Date: time.Now().Unix(),
	}

	pc, file, line, ok := runtime.Caller(2)
	d := runtime.FuncForPC(pc)

	if ok && d != nil {
		if lg.flag&(Lshortfile|Llongfile) != 0 {
			if lg.flag&Lshortfile != 0 {
				short := file
				for i := len(file) - 1; i > 0; i-- {
					if file[i] == '/' {
						short = file[i+1:]
						break
					}
				}
				file = short
			}
		}

		m.File = file
		m.Line = line
		m.Fnct = d.Name()

	} else {
		m.File = "???"
		m.Line = 0
	}

	return m
}

// Добавление параметров входящего запроса
func (m *Message) Query(obj interface{}) *Message {
	if qry, err := json.Marshal(&obj); err == nil {
		m.Qry = qry
	}
	return m
}

// Вывод лога в поток вывода
func (m *Message) Out() {
	lg.logOut(m)
}

// Отправка лога в буферизированный канал с последующим сохранением
func (m *Message) Save() {
	select {
	case <-lg.Ctx.Done():
		// Debug("  send Done err: %s \n", lg.Ctx.Err())
	case lg.ChMsg <- m:
		// Debug("send:%v len:%d \n", d, len(lg.ChData))
	}
}

// --------------------------------
//    Log
// --------------------------------

// Формирование сообщения на основе текстового сообщения.
func Msg(msg string) *Message {
	return newMessage(msg)
}

// Формирование сообщения о ошибке.
func Err(err error) *Message {
	if err == nil {
		return nil
	}
	return newMessage(err.Error())
}

// Эквивалентна выполению logStd(), с последующим вызовом os.Exit(1).
func Fatal(err error) {
	lg.logOut(newMessage(err.Error()))
	os.Exit(1)
}

// Форматированный вывод отладочной информации.
func Debug(format string, args ...interface{}) {
	fmt.Printf(format+" \n", args...)
}
