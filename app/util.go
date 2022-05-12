package app

import (
	"fmt"
	"strconv"
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
