package main

import (
    "github.com/philipparndt/go-logger"
    "os"
    "os/signal"
    "rnd7/edgeswitch-mqtt/config"
    "rnd7/edgeswitch-mqtt/edgeswitch"
    "rnd7/edgeswitch-mqtt/mqtt"
    "syscall"
    "time"
)

func mainLoop(cfg config.Config) {
    edgeswitch.Configure(cfg)

    for {
        edgeswitch.Execute()
        time.Sleep(time.Minute * 10)
    }
}

func main() {
    if len(os.Args) < 2 {
        logger.Error("No config file specified")
        os.Exit(1)
    }

    configFile := os.Args[1]
    logger.Info("Config file", configFile)
    cfg, err := config.LoadConfig(configFile)
    if err != nil {
        logger.Error("Failed loading config", err)
        return
    }

    logger.SetLevel(cfg.LogLevel)
    mqtt.Start(cfg.MQTT, edgeswitch.OnMessage)

    go mainLoop(cfg)

    logger.Info("Application is now ready. Press Ctrl+C to quit.")

    quitChannel := make(chan os.Signal, 1)
    signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
    <-quitChannel

    logger.Info("Received quit signal")
}
