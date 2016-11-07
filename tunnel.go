package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/solo-tunnel/sshconfig"
	"github.com/nitrous-io/solo-tunnel/sshtunnel"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Start SSH tunneling",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i",
					Usage: "path to SSH private key",
				},
				cli.StringSliceFlag{
					Name:  "p",
					Usage: "port mapping",
				},
			},
			Action: func(c *cli.Context) {
				if c.NArg() < 1 {
					logrus.Fatal("Please specify SSH URL.")
				}

				sshUrl := c.Args().Get(0)
				privateKeyPath := c.String("i")
				portPairs := c.StringSlice("p")

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
	}

	app.Run(os.Args)
}
