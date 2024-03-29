package mqtt

import (
    "github.com/philipparndt/mqtt-gateway/config"
    "github.com/philipparndt/mqtt-gateway/mqtt"
)

func Start(config config.MQTTConfig, onMessage mqtt.OnMessageListener) {
    mqtt.Start(config, "es_mqtt")
    mqtt.Subscribe(config.Topic + "/ports/+/poe/set", onMessage)
}

