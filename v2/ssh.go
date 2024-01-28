package main

import (
    "fmt"
)

var IsLogDebug = true

func RunCommands(user, password, ipPort string, cmds ...string) (string, error) {
    sessionKey := user + "_" + password + "_" + ipPort
    sessionManager.LockSession(sessionKey)
    defer sessionManager.UnlockSession(sessionKey)

    sshSession, err := sessionManager.GetSession(user, password, ipPort, "")
    if err != nil {
        LogError("GetSession error:%s", err)
        return "", err
    }

    LogDebug("=====================")
    LogDebug("=====================")
    LogDebug("=====================")
    LogDebug("=====================")

    for _, cmd := range cmds {
        LogDebug("WriteChannel CMD: ", cmd)
        sshSession.Write(cmd)
        var r = sshSession.ReadChannelTiming()
        LogDebug("ReadChannelTiming result: ", r)
    }

    LogDebug("=====================")
    LogDebug("=====================")
    LogDebug("=====================")
    LogDebug("=====================")

    //sshSession.WriteChannel(cmds...)
    //result := sshSession.ReadChannelTiming(2 * time.Second)
    //filteredResult := filterResult(result, cmds[0])
    return "", nil
}

func LogDebug(format string, a ...interface{}) {
    if IsLogDebug {
        fmt.Println("[DEBUG]:" + fmt.Sprintf(format, a...))
    }
}

func LogError(format string, a ...interface{}) {
    fmt.Println("[ERROR]:" + fmt.Sprintf(format, a...))
}
