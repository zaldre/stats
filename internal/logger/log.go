package logger

import (
	"fmt"
	"time"
)

var logLevel string

func SetLogLevel(level string) {
	logLevel = level
}

func OutLog(text string, obj interface{}) {

	if logLevel == "None" {
		return
	}
	timestamp := time.Now().Format("02.01.2006 15:04:05:")
	if text != "" {
		fmt.Printf("%s %s\n", timestamp, text)
	}
	if obj != nil {
		fmt.Println(timestamp)
		fmt.Println(obj)
	}
}
