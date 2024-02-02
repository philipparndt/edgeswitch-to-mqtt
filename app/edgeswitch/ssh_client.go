/**
 * Parts of this code are based on the work of
 * https://github.com/shenbowei/switch-ssh-go
 */

package edgeswitch

import (
    "golang.org/x/crypto/ssh"
    "net"
    "rnd7/edgeswitch-mqtt/logger"
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
       logger.Error("Cannot create session connection", err)
       return nil, err
   }
   if err := sshSession.muxShell(); err != nil {
       logger.Error("muxShell failed", err)
       return nil, err
   }
   if err := sshSession.start(); err != nil {
       logger.Error("Session start failed", err)
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
       logger.Error("SSH Dial failed", err)
       return err
   }
   session, err := client.NewSession()
   if err != nil {
       logger.Error("NewSession failed", err)
       return err
   }
   this.session = session
   return nil
}

func (this *SSHSession) muxShell() error {
   defer func() {
       if err := recover(); err != nil {
           logger.Error("SSHSession muxShell failed", err)
       }
   }()

   modes := ssh.TerminalModes{
       ssh.ECHO: 1, // no echo
       ssh.TTY_OP_ISPEED: 14_400,
       ssh.TTY_OP_OSPEED: 14_400,
   }

   if err := this.session.RequestPty("vt100", 80, 40, modes); err != nil {
       logger.Error("RequestPty failed:%s", err)
       return err
   }

   w, err := this.session.StdinPipe()
   if err != nil {
       logger.Error("StdinPipe() failed", err)
       return err
   }

   r, err := this.session.StdoutPipe()
   if err != nil {
       logger.Error("StdoutPipe() failed:%s", err)
       return err
   }

   in := make(chan string, 1024)
   out := make(chan string, 1024)
   go func() {
       defer func() {
           if err := recover(); err != nil {
               logger.Error("Goroutine muxShell write failed", err)
           }
       }()
       for cmd := range in {
           _, err := w.Write([]byte(cmd + "\n"))
           if err != nil {
               if err.Error() == "EOF" {
                   time.Sleep(time.Millisecond * 5)
                   return
               }
               logger.Error("Writer write failed", err)
               return
           }
       }
   }()

   go func() {
       defer func() {
           if err := recover(); err != nil {
               logger.Error("Goroutine muxShell read failed", err)
           }
       }()
       var (
           buf [65 * 1024]byte
           t   int
       )
       for {
           n, err := r.Read(buf[t:])
           if err != nil {
               if err.Error() == "EOF" {
                   time.Sleep(time.Millisecond * 5)
                   return
               }
               logger.Error("Reader read failed", err)
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
       logger.Error("Start shell error:%s", err)
       return err
   }

   return nil
}

func (this *SSHSession) Close() {
   defer func() {
       if err := recover(); err != nil {
           logger.Error("SSHSession Close failed", err)
       }
   }()
   if err := this.session.Close(); err != nil {
       logger.Error("Close session failed", err)
   }
   close(this.in)
   close(this.out)
}

func (this *SSHSession) Write(cmd string) {
   this.in <- cmd
}

func (this *SSHSession) ReadChannelData() string {
   result := this.readChannelData()
   for i := 0; i < 3000; i++ {
       time.Sleep(time.Millisecond * 100)
       var data = this.readChannelData()
       if data == "" || data == " " {
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
