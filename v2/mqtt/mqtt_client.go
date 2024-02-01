package mqtt

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "os"
    "rnd7/edgeswitch-mqtt/config"
    "rnd7/edgeswitch-mqtt/logger"
    "sync"
    "time"

    PAHO "github.com/eclipse/paho.mqtt.golang"
)

func onMessageReceived(client PAHO.Client, message PAHO.Message) {
    fmt.Printf("Received message on topic: %s\n", message.Topic())
    fmt.Printf("Message: %s\n", message.Payload())
}

func generateRandomClientID(length int) string {
    charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
    result := make([]byte, length)
    for i := range result {
        result[i] = charset[seededRand.Intn(len(charset))]
    }
    return "es_mqtt_" + string(result)
}

var client PAHO.Client
var baseTopic string

var connectionWg sync.WaitGroup

func Start(config config.MQTTConfig) {
    connectionWg.Add(1)
    go connect(config)
    connectionWg.Wait()
}

func connect(config config.MQTTConfig) {
    baseTopic = config.Topic
    statusTopic := baseTopic + "/bridge/state"
    clientID := generateRandomClientID(10)
    logger.Debug("Generated client ID: %s\n", clientID)

    opts := PAHO.NewClientOptions().
        AddBroker(config.URL).
        SetClientID(clientID).
        SetWill(statusTopic, "offline", 1, true)

    client = PAHO.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        logger.Error("Error connecting to MQTT broker:", token.Error())
        os.Exit(1)
    }
    defer client.Disconnect(250)

    PublishAbsolute(statusTopic, "online")

    logger.Info("Connected to MQTT broker", config.URL)
    connectionWg.Done()

    // Keep the connection active until the application is terminated
    select {}
}

func PublishAbsolute(topic string, message string) {
    token := client.Publish(topic, 0, false, message)
    token.Wait()

    logger.Debug("Published message", topic, message)

    if token.Error() != nil {
        logger.Error("Error publishing message", token.Error())
    }
}

func PublishJSON(topic string, data any) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        logger.Error("Error marshaling to JSON", err)
    } else {
        PublishAbsolute(baseTopic + "/" + topic, string(jsonData))
    }

}
