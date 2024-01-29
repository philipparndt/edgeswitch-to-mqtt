package config

import (
    "encoding/json"
    "os"
    "rnd7/edgeswitch-mqtt/logger"
)

type Config struct {
    MQTT struct {
        URL    string `json:"url"`
        Retain bool   `json:"retain"`
        Topic  string `json:"topic"`
        QoS    int    `json:"qos"`
    } `json:"mqtt"`
    EdgeSwitch struct {
        IP       string `json:"ip"`
        Username string `json:"username"`
        Password string `json:"password"`
        Ports    []struct {
            Name string `json:"name"`
            Port string `json:"port"`
        } `json:"ports"`
    } `json:"edgeswitch"`
}

func LoadConfig(file string) (Config, error) {
    data, err := os.ReadFile(file)
    if err != nil {
        logger.Error("Error reading config file", err)
        return Config{}, err
    }

    // Create a Config object
    var config Config

    // Unmarshal the JSON data into the Config object
    err = json.Unmarshal(data, &config)
    if err != nil {
        logger.Error("Unmarshalling JSON:", err)
        return Config{}, err
    }

    return config, nil
}
