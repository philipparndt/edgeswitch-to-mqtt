package main

import (
	"fmt"
    "os"
    "os/signal"
    "rnd7/edgeswitch-mqtt/config"
    "rnd7/edgeswitch-mqtt/edgeswitch"
    "rnd7/edgeswitch-mqtt/logger"
    "syscall"
    "time"
)

func mainLoop(cfg config.Config) {
    for {
        edgeswitch.Execute(cfg)
        time.Sleep(time.Minute)
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: ./app config.json")
        os.Exit(1)
    }

    configFile := os.Args[1]
    fmt.Println("Config file:", configFile)
    cfg, err := config.LoadConfig(configFile)
    if err != nil {
        logger.Error("Failed loading config", err)
        fmt.Println("Error loading config:", err)
        return
    }

    go mainLoop(cfg)

    quitChannel := make(chan os.Signal, 1)
    signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
    <-quitChannel

    fmt.Println("Exiting...")
}
