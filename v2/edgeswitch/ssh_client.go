package edgeswitch

import (
    "fmt"
    "golang.org/x/crypto/ssh"
    "net"
    "time"
)

type SSHSession struct {
   session     *ssh.Session
   in          chan string
   out         chan string
}


func StartSession(user, password, ipPort string) (*SSHSession, error) {
   sshSession := new(SSHSession)
   if err := sshSession.createConnection(user, password, ipPort); err != nil {
       fmt.Printf("[ERROR] cannot create session connection: %s", err.Error())
       return nil, err
   }
   if err := sshSession.muxShell(); err != nil {
       fmt.Printf("[ERROR] muxShell failed: %s", err.Error())
       return nil, err
   }
   if err := sshSession.start(); err != nil {
       fmt.Printf("[ERROR] session start failed: %s", err.Error())
       return nil, err
   }
   return sshSession, nil
}

func (this *SSHSession) createConnection(user, password, ipPort string) error {
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
       fmt.Printf("[ERROR] SSH Dial err:%s", err.Error())
       return err
   }
   session, err := client.NewSession()
   if err != nil {
       fmt.Printf("[ERROR] NewSession err:%s", err.Error())
       return err
   }
   this.session = session
   return nil
}

func (this *SSHSession) muxShell() error {
   defer func() {
       if err := recover(); err != nil {
           fmt.Printf("[ERROR] SSHSession muxShell failed:%s", err)
       }
   }()

   modes := ssh.TerminalModes{
       ssh.ECHO: 1, // no echo
       ssh.TTY_OP_ISPEED: 28800,
       ssh.TTY_OP_OSPEED: 28800,
   }

   if err := this.session.RequestPty("vt100", 80, 40, modes); err != nil {
       fmt.Printf("[ERROR] RequestPty failed:%s", err)
       return err
   }

   w, err := this.session.StdinPipe()
   if err != nil {
       fmt.Printf("[ERROR] StdinPipe() failed:%s", err.Error())
       return err
   }

   r, err := this.session.StdoutPipe()
   if err != nil {
       fmt.Printf("[ERROR] StdoutPipe() failed:%s", err.Error())
       return err
   }

   in := make(chan string, 1024)
   out := make(chan string, 1024)
   go func() {
       defer func() {
           if err := recover(); err != nil {
               fmt.Printf("[ERROR] Goroutine muxShell write err:%s", err)
           }
       }()
       for cmd := range in {
           _, err := w.Write([]byte(cmd + "\n"))
           if err != nil {
               fmt.Printf("[ERROR] Writer write err:%s", err.Error())
               return
           }
       }
   }()

   go func() {
       defer func() {
           if err := recover(); err != nil {
               fmt.Printf("[ERROR] Goroutine muxShell read err:%s", err)
           }
       }()
       var (
           buf [65 * 1024]byte
           t   int
       )
       for {
           n, err := r.Read(buf[t:])
           if err != nil {
               fmt.Printf("[ERROR] Reader read err:%s", err.Error())
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
       fmt.Printf("[ERROR] Start shell error:%s", err.Error())
       return err
   }

   return nil
}

func (this *SSHSession) Close() {
   defer func() {
       if err := recover(); err != nil {
           fmt.Printf("[ERROR] SSHSession Close err:%s", err)
       }
   }()
   if err := this.session.Close(); err != nil {
       fmt.Printf("[ERROR] Close session err:%s", err.Error())
   }
   close(this.in)
   close(this.out)
}

func (this *SSHSession) Write(cmd string) {
   this.in <- cmd
}

//func (this *SSHSession) ReadChannelExpect(timeout time.Duration, expects ...string) string {
//   output := ""
//   isDelayed := false
//   for i := 0; i < 300; i++ {
//       time.Sleep(time.Millisecond * 100)
//       newData := this.readChannelData()
//       if newData != "" {
//           output += newData
//           isDelayed = false
//           continue
//       }
//       for _, expect := range expects {
//           if strings.Contains(output, expect) {
//               return output
//           }
//       }
//       if !isDelayed {
//           time.Sleep(timeout)
//           isDelayed = true
//       } else {
//           return output
//       }
//   }
//   return output
//}

func (this *SSHSession) ReadChannelData() string {
   result := this.readChannelData()
   readMore := true
   for i := 0; i < 3000; i++ {
       time.Sleep(time.Millisecond * 10)
       var data = this.readChannelData()
       if data == "" {
           if readMore {
               readMore = false
               this.Write(" ") // Send a space to the device to get more data
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
