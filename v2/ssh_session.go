package main

import (
    "golang.org/x/crypto/ssh"
    "net"
    "strings"
    "time"
)

type SSHSession struct {
    session     *ssh.Session
    in          chan string
    out         chan string
    brand       string
    lastUseTime time.Time
}

func NewSSHSession(user, password, ipPort string) (*SSHSession, error) {
    sshSession := new(SSHSession)
    if err := sshSession.createConnection(user, password, ipPort); err != nil {
        LogError("NewSSHSession createConnection error:%s", err.Error())
        return nil, err
    }
    if err := sshSession.muxShell(); err != nil {
        LogError("NewSSHSession muxShell error:%s", err.Error())
        return nil, err
    }
    if err := sshSession.start(); err != nil {
        LogError("NewSSHSession start error:%s", err.Error())
        return nil, err
    }
    sshSession.lastUseTime = time.Now()
    sshSession.brand = ""
    return sshSession, nil
}

func (this *SSHSession) GetLastUseTime() time.Time {
    return this.lastUseTime
}

func (this *SSHSession) UpdateLastUseTime() {
    this.lastUseTime = time.Now()
}

func (this *SSHSession) createConnection(user, password, ipPort string) error {
    LogDebug("<Test> Begin connect")
    client, err := ssh.Dial("tcp", ipPort, &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.Password(password),
        },
        HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
            return nil
        },
        Timeout: 20 * time.Second,
        Config: ssh.Config{
            Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com",
                "arcfour256", "arcfour128", "aes128-cbc", "aes256-cbc", "3des-cbc", "des-cbc",
            },
        },
    })
    if err != nil {
        LogError("SSH Dial err:%s", err.Error())
        return err
    }
    LogDebug("<Test> End connect")
    LogDebug("<Test> Begin new session")
    session, err := client.NewSession()
    if err != nil {
        LogError("NewSession err:%s", err.Error())
        return err
    }
    this.session = session
    LogDebug("<Test> End new session")
    return nil
}

func (this *SSHSession) muxShell() error {
    defer func() {
        if err := recover(); err != nil {
            LogError("SSHSession muxShell err:%s", err)
        }
    }()
    modes := ssh.TerminalModes{
        ssh.ECHO:          1,     // disable echoing
        ssh.TTY_OP_ISPEED: 28800, // 14400, // input speed = 14.4kbaud
        ssh.TTY_OP_OSPEED: 28800, // output speed = 14.4kbaud
    }
    if err := this.session.RequestPty("vt100", 250, 80, modes); err != nil {
        LogError("RequestPty error:%s", err)
        return err
    }
    w, err := this.session.StdinPipe()
    if err != nil {
        LogError("StdinPipe() error:%s", err.Error())
        return err
    }
    r, err := this.session.StdoutPipe()
    if err != nil {
        LogError("StdoutPipe() error:%s", err.Error())
        return err
    }

    in := make(chan string, 1024)
    out := make(chan string, 1024)
    go func() {
        defer func() {
            if err := recover(); err != nil {
                LogError("Goroutine muxShell write err:%s", err)
            }
        }()
        for cmd := range in {
            _, err := w.Write([]byte(cmd + "\n"))
            if err != nil {
                LogDebug("Writer write err:%s", err.Error())
                return
            }
        }
    }()

    go func() {
        defer func() {
            if err := recover(); err != nil {
                LogError("Goroutine muxShell read err:%s", err)
            }
        }()
        var (
            buf [65 * 1024]byte
            t   int
        )
        for {
            n, err := r.Read(buf[t:])
            if err != nil {
                LogDebug("Reader read err:%s", err.Error())
                return
            }
            t += n
            out <- string(buf[:t])
            t = 0
        }
    }()
    this.in = in
    this.out = out
    return nil
}

func (this *SSHSession) start() error {
    if err := this.session.Shell(); err != nil {
        LogError("Start shell error:%s", err.Error())
        return err
    }

    // this.ReadChannelExpect(time.Second, "#", ">", "]")
    return nil
}

func (this *SSHSession) CheckSelf() bool {
    //defer func() {
    //    if err := recover(); err != nil {
    //        LogError("SSHSession CheckSelf err:%s", err)
    //    }
    //}()
    //
    //this.WriteChannel("\n")
    //result := this.ReadChannelExpect(2*time.Second, "#", ">", "]")
    //if strings.Contains(result, "#") ||
    //    strings.Contains(result, ">") ||
    //    strings.Contains(result, "]") {
    //    return true
    //}
    return true
}

func (this *SSHSession) Close() {
    defer func() {
        if err := recover(); err != nil {
            LogError("SSHSession Close err:%s", err)
        }
    }()
    if err := this.session.Close(); err != nil {
        LogError("Close session err:%s", err.Error())
    }
    close(this.in)
    close(this.out)
}

func (this *SSHSession) WriteChannel(cmds ...string) {
    LogDebug("WriteChannel <cmds=%v>", cmds)
    for _, cmd := range cmds {
        LogDebug("WriteChannel CMD: ", cmd)
        this.in <- cmd
    }
}

func (this *SSHSession) Write(cmd string) {
    this.in <- cmd
}


func (this *SSHSession) ReadChannelExpect(timeout time.Duration, expects ...string) string {
    LogDebug("ReadChannelExpect <wait timeout = %d>", timeout/time.Millisecond)
    output := ""
    isDelayed := false
    for i := 0; i < 300; i++ { //最多从设备读取300次，避免方法无法返回
        time.Sleep(time.Millisecond * 100) //每次睡眠0.1秒，使out管道中的数据能积累一段时间，避免过早触发default等待退出
        newData := this.readChannelData()
        LogDebug("ReadChannelExpect: read chanel buffer: %s", newData)
        if newData != "" {
            output += newData
            isDelayed = false
            continue
        }
        for _, expect := range expects {
            if strings.Contains(output, expect) {
                return output
            }
        }
        if !isDelayed {
            LogDebug("ReadChannelExpect: delay for timeout")
            time.Sleep(timeout)
            isDelayed = true
        } else {
            return output
        }
    }
    return output
}

func (this *SSHSession) ReadChannelTiming() string {
    result := this.readChannelData()
    readMore := true
    for i := 0; i < 3000; i++ {
        time.Sleep(time.Millisecond * 10)
        var data = this.readChannelData()
        if data == "" {
            if readMore {
                readMore = false
                this.Write(" ")  // Read more data
                time.Sleep(time.Millisecond * 100)
                continue
            }
            // No more data in the channel
            break
        }

        result += data
    }

    return result
}

func (this *SSHSession) readChannelData() string {
    output := ""
    for {
        select {
        case channelData, ok := <-this.out:
            if !ok {
                return output
            }
            output += channelData
        default:
            return output
        }
    }
}
