package edgeswitch

import (
    "fmt"
    "rnd7/edgeswitch-mqtt/config"
)

func Execute(config config.Config) {
    user := config.EdgeSwitch.Username
    password := config.EdgeSwitch.Password
    ipPort := config.EdgeSwitch.IP + ":22"

    //cmds := make([]string, 0)
    //cmds = append(cmds, "configure")
    //cmds = append(cmds, "show interface ethernet all")

    session, err := StartSession(user, password, ipPort)
    if err != nil {
        fmt.Println("[ERROR] StartSession:\n", err.Error())
        return
    }

    session.Write("configure")
    var _ = session.ReadChannelData()

    session.Write("show interface ethernet all")
    channelData, err := ParseChannelData(session.ReadChannelData())
    if err != nil {
        return
    }

    fmt.Println("Data:", channelData)
    fmt.Println("------------------------------------------------------------------------")
    for _, port := range config.EdgeSwitch.Ports {
        session.Write("show poe status " + port.Port)
        info, err := ParseDeviceInfo(session.ReadChannelData())
        if err != nil {
            return
        }
        fmt.Println("Info:", info)
    }
}
