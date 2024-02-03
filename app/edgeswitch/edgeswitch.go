package edgeswitch

import (
    "rnd7/edgeswitch-mqtt/config"
    "rnd7/edgeswitch-mqtt/logger"
    "rnd7/edgeswitch-mqtt/mqtt"
    "strings"
)

var cfg config.Config
var portToName = make(map[string]string)
var nameToPort map[string]string

func Configure(configuration config.Config) {
    cfg = configuration

    for _, portInfo := range cfg.EdgeSwitch.Ports {
        portToName[portInfo.Port] = portInfo.Name
    }

    nameToPort = make(map[string]string)
    for _, portInfo := range configuration.EdgeSwitch.Ports {
        nameToPort[portInfo.Name] = portInfo.Port
    }
}

func OnMessage(topic string, payloadData []byte) {
    payload := string(payloadData)
    paths := strings.Split(topic, "/")
    name := strings.TrimSpace(paths[len(paths) - 3])
    port := nameToPort[name]

    if port == "" {
        logger.Error("Port not found for name", name)
        return
    }

    if strings.EqualFold(payload, "true") {
        TurnPoEOn(port)
    } else if strings.EqualFold(payload, "false") {
        TurnPoEOff(port)
    }
}

func TurnPoEOff(port string) {
    logger.Info("Switching PoE OFF for port", port)
    poeSwitch(port, "shutdown")
}

func TurnPoEOn(port string) {
    logger.Info("Switching PoE ON for port", port)
    poeSwitch(port, "auto")
}

func poeSwitch(port string, opmode string) {
    execute([]string{
        "configure",
        "interface " + port,
        "poe opmode " + opmode})
}

func execute(commands []string) {
    logger.Debug("Executing commands", commands)
    session, err := StartSession(cfg.EdgeSwitch.Username, cfg.EdgeSwitch.Password, cfg.EdgeSwitch.IP + ":22")
    if err != nil {
        logger.Error("StartSession failed", err)
        return
    }
    logger.Debug("Session started")

    defer session.Close()

    for _, cmd := range commands {
        session.Write(cmd)
        var x = session.ReadChannelData()
        logger.Debug("Command:", cmd, "Response:", x)
    }
}

func Execute() {
    logger.Debug("Updating...")

    session, err := StartSession(cfg.EdgeSwitch.Username, cfg.EdgeSwitch.Password, cfg.EdgeSwitch.IP + ":22")
    if err != nil {
        logger.Error("StartSession failed", err)
        return
    }

    defer session.Close()

    // Disable pagination
    session.Write("terminal length 0")
    var _ = session.ReadChannelData()

    session.Write("configure")
    var _ = session.ReadChannelData()

    session.Write("show interface ethernet all")
    channelData, err := ParseChannelData(session.ReadChannelData())
    if err != nil {
        return
    }

    for _, chdata := range channelData {
        mqtt.PublishJSON("ports/" + toName(portToName, chdata.Port) + "/transmit", mqtt.TransmitMessage{
            Port: chdata.Port,
            BytesTx: chdata.BytesTx,
            BytesRx: chdata.BytesRx,
            PacketsTx: chdata.PacketsTx,
            PacketsRx: chdata.PacketsRx,
            TotalBytes: chdata.BytesTx + chdata.BytesRx,
        })
    }

    aggregated := mqtt.AggregatedEnergy{
        EnergySum: 0.0,
        WhrSum: 0.0,
    }

    for _, port := range cfg.EdgeSwitch.Ports {
        session.Write("show poe status " + port.Port)
        info, err := ParseDeviceInfo(session.ReadChannelData())
        if err != nil {
            return
        }

        for _, deviceInfo := range info {
            mqtt.PublishJSON("ports/" + toName(portToName, deviceInfo.Intf) + "/poe", mqtt.DeviceDataMessage{
                Interface:   deviceInfo.Intf,
                Detection:   deviceInfo.Detection,
                Status:      convertStatus(deviceInfo.Detection),
                Class:       deviceInfo.Class,
                Energy:      deviceInfo.ConsumedW,
                Voltage:     deviceInfo.VoltageV,
                CurrentmA:   deviceInfo.CurrentmA,
                TotalWhr:    deviceInfo.ConsumedMeterWhr,
                Temperature: deviceInfo.TemperatureC,
            })

            aggregated.EnergySum += deviceInfo.ConsumedW
            aggregated.WhrSum += deviceInfo.ConsumedMeterWhr
        }
    }

    // round to 2 decimal places
    aggregated.EnergySum = float64(int(aggregated.EnergySum * 100)) / 100
    aggregated.WhrSum = float64(int(aggregated.WhrSum * 100)) / 100

    mqtt.PublishJSON("aggregated", aggregated)

    logger.Debug("Update completed")
}


func convertStatus(statusString string) int {
    // Compare the string case-insensitively
    if strings.EqualFold(statusString, "Good") {
        return 1
    } else {
        return 0
    }
}

func toName(portToName map[string]string, port string) string {
    name := portToName[port]
    if name == "" {
        name = port
        name = strings.ReplaceAll(name, "/", "_")
    }
    return name

}
