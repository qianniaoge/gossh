package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"gossh/config"
	"gossh/sshclient"
)

const example = `
Example <hosts.conf>:
[profile]
username = user
password = pass
identityFile = ~/.ssh/id_rsa
identityPass = "Null or private key's password"
port = 22
parallel = 10

[hosts]
192.168.1.3-20
192.168.1.100
192.168.1.253
`

const summary  = `Summary:
All Executed Hosts: %d
Success:            %d
Failures:           %d
Failed Hosts:
`

type Counter struct {
	sync.Mutex
	Data int
	Hosts []string
}

func (c *Counter) Incre(host string)  {
	c.Lock()
	defer c.Unlock()
	c.Data++
	c.Hosts[c.Data-1] = host
}

func (c *Counter) Decre()  {
	c.Lock()
	defer c.Unlock()
	c.Data--
}


func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Usage: ", os.Args[0], "command")
		return
	}

	var (
		wg    sync.WaitGroup
		parallel int
	)

	cmd := strings.Join(os.Args[1:], " ")
	conf, err := config.LoadConfig("hosts.conf")
	if err != nil {
		fmt.Println(err, example)
		return
	}

	success := &Counter{Hosts: make([]string, len(conf.Hosts))}
	fail := &Counter{Hosts: make([]string, len(conf.Hosts))}

	for _, host := range conf.Hosts {
		parallel++
		wg.Add(1)
		go func(host string) {
			ssh := sshclient.NewSSH(
				host, conf.User,
				conf.Password,
				conf.IdentityFile,
				conf.IdentityPass,
				conf.Port,
				)
			if ret := ssh.PrintRun(cmd); ret == 0 {
				success.Incre(host)
			} else {
				fail.Incre(host)
			}
			wg.Done()
		}(host)

		if parallel == conf.Parallel {
			wg.Wait()
			parallel = 0
		}
	}
	wg.Wait()

	fmt.Printf(summary, success.Data+fail.Data, success.Data, fail.Data)

	for _, v := range fail.Hosts{
		if v != "" {
			fmt.Printf("     %s\n", v)
		}
	}
}
