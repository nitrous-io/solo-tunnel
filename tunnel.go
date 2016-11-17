package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/solo-tunnel/sshconfig"
	"github.com/nitrous-io/solo-tunnel/sshtunnel"
	"github.com/urfave/cli"
)

var Version string

func main() {
	app := cli.NewApp()
	app.Usage = "SSH tunneling tool for Nitrous Solo"
	app.Version = Version
	app.UsageText = "solo-tunnel command <SSH URL> [command options]"
	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Start tunneling to Nitrous Solo container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i",
					Usage: "Path to SSH private key [Default: $HOME/.ssh/id_rsa or $HOME/.ssh/id_dsa]",
				},
				cli.StringSliceFlag{
					Name:  "p",
					Usage: "Port mappings to Nitrous Solo container. e.g 3000:4000 [Default: 6877:6877]",
				},
			},
			Action: func(c *cli.Context) {
				if c.NArg() < 1 {
					logrus.Fatal("Please specify SSH URL. e.g ssh://nitrous@1.2.3.4:22")
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
