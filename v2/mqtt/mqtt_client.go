package mqtt

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "os"
    "rnd7/edgeswitch-mqtt/config"
    "rnd7/edgeswitch-mqtt/logger"
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
    return string(result)
}

var client PAHO.Client
var baseTopic string

func Connect(config config.MQTTConfig) {
    baseTopic = config.Topic
    statusTopic := baseTopic + "/bridge/state"
    clientID := generateRandomClientID(15)
    fmt.Printf("Generated client ID: %s\n", clientID)

    opts := PAHO.NewClientOptions().
        AddBroker(config.URL).
        SetClientID(clientID).
        SetWill(statusTopic, "offline", 1, true)

    client = PAHO.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        fmt.Println("Error connecting to MQTT broker:", token.Error())
        os.Exit(1)
    }
    defer client.Disconnect(250)

    PublishAbsolute(statusTopic, "online")

    // Keep the connection active until the application is terminated
    select {}
}

func PublishAbsolute(topic string, message string) {
    logger.Info("Publishing message to topic:", topic, message)
    token := client.Publish(topic, 0, false, message)
    token.Wait()

    if token.Error() != nil {
        logger.Error("Error publishing message", token.Error())
    } else {
        logger.Info("Message published successfully")
    }
}

func Publish(topic string, message string) {
    PublishAbsolute(baseTopic + "/" + topic, message)
}

func PublishJSON(topic string, data any) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        logger.Error("Error marshaling to JSON", err)
    } else {
        PublishAbsolute(baseTopic + "/" + topic, string(jsonData))
    }

}
