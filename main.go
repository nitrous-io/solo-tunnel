package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/go-tunnel/sshconfig"
	"github.com/nitrous-io/go-tunnel/tunnel"
)

func main() {
	sshHost := os.Args[1]
	privateKeyPath := os.Args[2]
	portPairs := os.Args[3]

	ports, _ := tunnel.ParsePort(portPairs)
	host, config, _ := sshconfig.Config(sshHost, privateKeyPath)

	done := make(chan bool)
	logrus.SetLevel(logrus.WarnLevel)
	tunnel.NewSSHTunnel(host, config, ports).Start()
	<-done
}
