package logger

import (
    "fmt"
    "time"
)

func Error(message string, a ...any) {
    isoDateTime := time.Now().Format("2006-01-02 15:04:05")
    fmt.Println("[ERROR]", isoDateTime, message, a)
}

func Info(message string, a ...any) {
    isoDateTime := time.Now().Format("2006-01-02 15:04:05")

    if len(a) == 0 {
        fmt.Println("[INFO]", isoDateTime, message)
        return
    }
    fmt.Println("[INFO]", isoDateTime, message, a)
}


func Err(message string, err any) {
    fmt.Printf("[ERROR] %s: %s", message, err)
}
