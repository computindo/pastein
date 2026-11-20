package logger

import (
	"fmt"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorGray   = "\033[90m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
)

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Info(context, message string) {
	fmt.Printf("%s[%s]%s %s[INFO]%s  %s[%-8s]%s %s\n",
		colorGray, timestamp(), colorReset,
		colorGreen, colorReset,
		colorCyan, context, colorReset,
		message,
	)
}

func Error(context, message string) {
	fmt.Printf("%s[%s]%s %s[ERROR]%s %s[%-8s]%s %s\n",
		colorGray, timestamp(), colorReset,
		colorRed, colorReset,
		colorCyan, context, colorReset,
		message,
	)
}

func Warn(context, message string) {
	fmt.Printf("%s[%s]%s %s[WARN]%s  %s[%-8s]%s %s\n",
		colorGray, timestamp(), colorReset,
		colorYellow, colorReset,
		colorCyan, context, colorReset,
		message,
	)
}
