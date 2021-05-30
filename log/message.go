package log

import (
	//"fmt"
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
	Msg  string `db:"message" json:"msg"`
	Qry  string `db:"query" json:"query"`
	Date int64  `db:"date" json:"date"`
}

//func logNew(msg string) *dbLog {
func newMessage(msg string) *Message {
	l := &Message{Msg: msg}

	pc, file, line, ok := runtime.Caller(2)
	d := runtime.FuncForPC(pc)
	if ok && d != nil {
		l.File = file
		l.Line = line
		l.Fnct = d.Name()
		l.Date = time.Now().Unix()
	}

	return l
}

// добавление параметров входящего запроса
func (d *Message) Query(q string) {
	d.Qry = q
}

// вывод лога в терминал
func (d *Message) Out() {
	logStd(d)
}

// отправка лога в буферизированный канал с последующим сохранением
func (d *Message) Save() {
	select {
	case <-lg.Ctx.Done():
		// Debug("  send Done err: %s \n", lg.Ctx.Err())
		//close(lg.ChData)
	case lg.ChData <- d:
		// Debug("send:%v len:%d \n", d, len(lg.ChData))
	}
}
