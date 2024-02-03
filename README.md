# edgeswitch-to-mqtt-gw

[![mqtt-smarthome](https://img.shields.io/badge/mqtt-smarthome-blue.svg)](https://github.com/mqtt-smarthome/mqtt-smarthome)

Maintain a topic with port status for an EdgeSwitch. 
Tested with EdgeSwitch 16 150W

Ability to turn on/off PoW power of the ports via MQTT.

# Messages

## Example message

Topic: `home/ip/switch/ports/wifi/poe`
```json
{
  "interface":"0/3",
  "detection":"Good",
  "status":1,
  "class":"Class0",
  "energy":3.25,
  "voltage":53.59,
  "current_mA":60.66,
  "total_Whr":667.27,
  "temperature":41
}
```

## Turn on/off PoW power

Post a message to `home/ip/switch/ports/wifi/poe/set` with the following payload: `true` or `false` 
to turn on/off the PoW power.

## Example configuration

```json
{
  "mqtt": {
    "url": "tcp://192.168.0.1:1883",
    "retain": true,
    "topic": "home/ip/switch",
    "qos": 2
  },
  "edgeswitch": {
    "ip": "192.168.1.2",
    "username": "ubnt",
    "password": "ubnt",
    "ports": [
      { "name": "router_port1", "port": "0/1" },
      { "name": "router_port4_guest", "port": "0/2" },
      { "name": "wifi", "port": "0/3" }
    ]
  }
}
```

# Bridge status

The bridge maintains a status topic:

## Topic: `.../bridge/state`

| Value     | Description                          |
| --------- | ------------------------------------ |
| `online`  | The bridge is started                |
| `offline` | The bridge is currently not started. |

# run

Copy the `config-example.json` to `/production/config/config.json`

```
cd ./production
docker-compose up -d
```

# build

run `go build .` in the `app` folder

# NodeJS version

The main version of this bridge is written in Go. 
There is also a NodeJS version available in the `0.x` branch (no longer maintained).
