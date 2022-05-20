package app

import (
	"fmt"
	"strconv"
	"time"
)

const (
	sTimer time.Duration = 5 * time.Minute
	mTimer time.Duration = 10 * time.Minute
	lTimer time.Duration = 20 * time.Minute
	xTimer time.Duration = 60 * time.Minute
)

// --------------------------------
//    Вопомогательный функционал
// --------------------------------

type H map[string]interface{}

func I16S(i uint16) string {
	return strconv.FormatUint(uint64(i), 10)
}

func I3S(i uint32) string {
	return strconv.FormatUint(uint64(i), 10)
}

func I6S(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func Debug(format string, args ...interface{}) {
	if !cfg.Debug {
		return
	}
	fmt.Printf(format+" \n", args...)
}
