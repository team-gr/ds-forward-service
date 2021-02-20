package domain

import (
	"fmt"
	"log"
	"runtime"
)

type Logger struct {}

func NewLogger() Logger {
	return Logger{}
}

func (l Logger) Info(layout string, args ...interface{}) {
	print("INFO", layout, args...)
}

func (l Logger) Warn(layout string, args ...interface{}) {
	print("WARN", layout, args...)
}

func (l Logger) Error(layout string, args ...interface{}) {
	print("ERROR", layout, args...)
}

func (l Logger) Debug(layout string, args ...interface{}) {
	print("DEBUG", layout, args...)
}


func print(level string, layout string, args ...interface{}) {
	_, fn, line, _ := runtime.Caller(2)
	log.Println(fmt.Sprintf("[%s] %v:%v ", level, fn, line) + fmt.Sprintf(layout, args...))
}
