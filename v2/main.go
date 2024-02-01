package main

import (
    "os"
    "os/signal"
    "rnd7/edgeswitch-mqtt/config"
    "rnd7/edgeswitch-mqtt/edgeswitch"
    "rnd7/edgeswitch-mqtt/logger"
    "rnd7/edgeswitch-mqtt/mqtt"
    "syscall"
    "time"
)

func mainLoop(cfg config.Config) {
    for {
        edgeswitch.Execute(cfg)
        time.Sleep(time.Minute * 5)
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

    mqtt.Start(cfg.MQTT)

    go mainLoop(cfg)

    quitChannel := make(chan os.Signal, 1)
    signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
    <-quitChannel

    logger.Info("Received quit signal")
}
