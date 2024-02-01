package edgeswitch

import (
    "rnd7/edgeswitch-mqtt/config"
    "rnd7/edgeswitch-mqtt/logger"
    "rnd7/edgeswitch-mqtt/mqtt"
    "strings"
)

func ConvertStatus(statusString string) int {
    // Compare the string case-insensitively
    if strings.EqualFold(statusString, "Good") {
        return 1
    } else {
        return 0
    }
}

func ToName(portToName map[string]string, port string) string {
    name := portToName[port]
    if name == "" {
        name = port
        name = strings.ReplaceAll(name, "/", "_")
    }
    return name

}

func Execute(config config.Config) {
    logger.Debug("Updating...")
    user := config.EdgeSwitch.Username
    password := config.EdgeSwitch.Password
    ipPort := config.EdgeSwitch.IP + ":22"

    portToName := make(map[string]string)
    for _, portInfo := range config.EdgeSwitch.Ports {
        portToName[portInfo.Port] = portInfo.Name
    }

    session, err := StartSession(user, password, ipPort)
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
        mqtt.PublishJSON(ToName(portToName, chdata.Port) + "/transmit", mqtt.TransmitMessage{
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

    for _, port := range config.EdgeSwitch.Ports {
        session.Write("show poe status " + port.Port)
        info, err := ParseDeviceInfo(session.ReadChannelData())
        if err != nil {
            return
        }

        for _, deviceInfo := range info {
            mqtt.PublishJSON(ToName(portToName, deviceInfo.Intf) + "/poe", mqtt.DeviceDataMessage{
                Interface: deviceInfo.Intf,
                Detection: deviceInfo.Detection,
                Status: ConvertStatus(deviceInfo.Detection),
                Class: deviceInfo.Class,
                Energy: deviceInfo.ConsumedW,
                Voltage: deviceInfo.VoltageV,
                CurrentmA: deviceInfo.CurrentmA,
                TotalWhr: deviceInfo.ConsumedMeterWhr,
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
