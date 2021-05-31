package log

import (
	//"fmt"
	"context"
	"os"
	"runtime"
	"time"

	"encoding/json"
	"sync/atomic"
	"unsafe"
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

// Формирование сообщения лога, для последующего сохранения
func newMessage(msg string) *Message {
	m := &Message{Msg: msg}

	pc, file, line, ok := runtime.Caller(2)
	d := runtime.FuncForPC(pc)
	if ok && d != nil {
		m.File = file
		m.Line = line
		m.Fnct = d.Name()
		m.Date = time.Now().Unix()
	}

	return m
}

// Добавление параметров входящего запроса
func (m *Message) Query(q string) {
	m.Qry = q
}

// Вывод лога в терминал
func (m *Message) Out() {
	m.logStd()
}

// Отправка лога в буферизированный канал с последующим сохранением
func (m *Message) Save() {
	select {
	case <-lg.Ctx.Done():
		// Debug("  send Done err: %s \n", lg.Ctx.Err())
		//close(lg.ChData)
	case lg.ChMsg <- m:
		// Debug("send:%v len:%d \n", d, len(lg.ChData))
	}
}

// --------------------------------
//    Log Writer
// --------------------------------

// Отправка записи сетевому сборщику.
func (m *Message) logNet() {
	// rpc.Dial()
	Debug("logNet [%s] line:%d file:%s \nerr:%s", m.Fnct, m.Line, m.File, m.Msg)
}

// Регистрация записи в базе данных Postgre.
func (m *Message) logDb() {
	Debug("logDb msg:%s", m.Msg)
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	_, err := lg.pdb.Exec(ctx, `INSERT INTO main.log(file, line, function, message, datecreate) VALUES ($1, $2, $3, $4, $5)`, m.File, m.Line, m.Fnct, m.Msg, m.Date)
	if IsDbErr(err) {
		m.logFile()
	}
}

// Регистрация записи в текстовом файле.
func (m *Message) logFile() {
	Debug("logFile msg:%s", m.Msg)
	fp := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&lg.file)))
	file := (*os.File)(fp)
	if err := json.NewEncoder(file).Encode(&m); err != nil {
		Fatal(err)
	}
}

// Вывод сообщения в терминал.
func (m *Message) logStd() {
	Debug("logStd [%s] line:%d file:%s \nerr:%s", m.Fnct, m.Line, m.File, m.Msg)
}
