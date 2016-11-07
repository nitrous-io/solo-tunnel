package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/go-tunnel/sshconfig"
	"github.com/nitrous-io/go-tunnel/sshtunnel"
)

func main() {
	sshUrl := os.Args[1]
	privateKeyPath := os.Args[2]
	portPairs := os.Args[3]

	ports, err := sshtunnel.ParsePort(portPairs)
	if err != nil {
		logrus.Fatal(err)
	}

	host, config, err := sshconfig.Config(sshUrl, privateKeyPath)
	if err != nil {
		logrus.Fatal(err)
	}

	done := make(chan bool)
	logrus.SetLevel(logrus.DebugLevel)
	if err := sshtunnel.NewSSHTunnel(host, config, ports).Start(); err != nil {
		logrus.Fatal(err)
	}
	<-done
}
