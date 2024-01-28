package logger

import "fmt"

func Error(message string, err error) {
    fmt.Printf("[ERROR] %s: %s", message, err.Error())
}

func Err(message string, err any) {
    fmt.Printf("[ERROR] %s: %s", message, err)
}
