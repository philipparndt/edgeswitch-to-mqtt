package logger

import (
    "log"
    "os"
    "strings"
    "time"
)

var CustomLogger *log.Logger

func init() {
    CustomLogger = log.New(os.Stdout, "", 0)
}

func Log(severity string, message string, a ...any) {
    if len(a) == 0 {
        CustomLogger.Printf("[%s] %s %s\n", strings.ToUpper(severity), time.Now().Format("2006-01-02T15:04:05"), message)
        return
    } else {
        CustomLogger.Printf("[%s] %s %s %s\n", strings.ToUpper(severity), time.Now().Format("2006-01-02T15:04:05"), message, a)
    }
}

func Info(message string, a ...any) {
    Log("info", message, a...)
}

func Error(message string, a ...any) {
    Log("error", message, a...)
}
