package main

import (
	"fmt"
    "rnd7/edgeswitch-mqtt/edgeswitch"
)

func main() {


	cmds := make([]string, 0)
	cmds = append(cmds, "configure")
	cmds = append(cmds, "show interface ethernet all")
    cmds = append(cmds, "show poe status 0/5")
    cmds = append(cmds, "show poe status 0/6")

    session, err := edgeswitch.StartSession(user, password, ipPort)
    if err != nil {
        fmt.Println("[ERROR] StartSession:\n", err.Error())
        return
    }

    for _, cmd := range cmds {
        session.Write(cmd)
        var r = session.ReadChannelData()
        fmt.Println("ReadChannelData result: ", r)
    }
}
