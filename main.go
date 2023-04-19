package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/nathan-osman/sensorpi/input/onewire"
	"github.com/nathan-osman/sensorpi/manager"
	_ "github.com/nathan-osman/sensorpi/output/influxdb"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "sensorpi",
		Usage: "monitor sensors connected to a Raspberry Pi",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Required: true,
				EnvVars:  []string{"CONFIG"},
				Usage:    "filename of configuration file",
			},
		},
		Commands: []*cli.Command{
			installCommand,
		},
		Action: func(c *cli.Context) error {

			// Create the manager from the config file
			m, err := manager.New(c.String("config"))
			if err != nil {
				return err
			}
			defer m.Close()

			// Wait for SIGINT or SIGTERM
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s\n", err.Error())
	}
}
