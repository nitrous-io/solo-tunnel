package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/solo-tunnel/sshconfig"
	"github.com/nitrous-io/solo-tunnel/sshtunnel"
	"github.com/nitrous-io/solo-tunnel/terminal"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Start SSH tunneling",
			Action: func(c *cli.Context) {
				if c.NArg() < 3 {
					logrus.Fatal("Not enough arguments.")
				}

				sshUrl := c.Args().Get(0)
				privateKeyPath := c.Args().Get(1)
				portPairs := c.Args().Get(2)

				ports, err := sshtunnel.ParsePort(portPairs)
				if err != nil {
					logrus.Fatal(err)
				}

				host, config, err := sshconfig.Config(sshUrl, privateKeyPath)
				if err != nil {
					logrus.Fatal(err)
				}

				done := make(chan struct{})
				logrus.SetLevel(logrus.DebugLevel)
				if err := sshtunnel.NewSSHTunnel(host, config, ports).Start(); err != nil {
					logrus.Fatal(err)
				}
				<-done
			},
		},
		{
			Name:  "shell",
			Usage: "Open shell via SSH",
			Action: func(c *cli.Context) {
				if c.NArg() < 2 {
					logrus.Fatal("Not enough arguments.")
				}

				sshUrl := c.Args().Get(0)
				privateKeyPath := c.Args().Get(1)

				logrus.Println("Start ssh", sshUrl, privateKeyPath)

				terminal.Start(sshUrl, privateKeyPath)
			},
		},
	}

	app.Run(os.Args)
}
