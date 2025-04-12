package log

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

var (
	errorColor   = color.New(color.FgRed).SprintFunc()
	warningColor = color.New(color.FgYellow).SprintFunc()
	infoColor    = color.New(color.FgBlue).SprintFunc()
	successColor = color.New(color.FgGreen).SprintFunc()
)

func getTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Error(msg string, err error) {
	fmt.Printf("[%s] %s (%v)\n", getTime(), errorColor("[ERROR] ")+msg, err)
}

func Warning(msg string) {
	fmt.Printf("[%s] %s\n", getTime(), warningColor("[WARNING] ")+msg)
}

func Message(msg string) {
	fmt.Printf("[%s] %s\n", getTime(), infoColor("[INFO] ")+msg)
}

func Success(msg string) {
	fmt.Printf("[%s] %s\n", getTime(), successColor("[SUCCESS] ")+msg)
}
