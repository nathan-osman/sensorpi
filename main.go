package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nathan-osman/sensorpi/manager"
	_ "github.com/nathan-osman/sensorpi/plugins/console"
	_ "github.com/nathan-osman/sensorpi/plugins/grove-moisture"
	_ "github.com/nathan-osman/sensorpi/plugins/influxdb"
	_ "github.com/nathan-osman/sensorpi/plugins/onewire"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

var configFlag = &cli.StringFlag{
	Name:    "config",
	Value:   "/etc/sensorpi/config.yaml",
	EnvVars: []string{"CONFIG"},
	Usage:   "filename of configuration file",
}

func main() {
	app := &cli.App{
		Name:  "sensorpi",
		Usage: "monitor sensors connected to a Raspberry Pi",
		Flags: []cli.Flag{
			configFlag,
			&cli.BoolFlag{
				Name:    "debug",
				EnvVars: []string{"DEBUG"},
				Usage:   "enable debug mode",
			},
		},
		Commands: []*cli.Command{
			installCommand,
		},
		Action: func(c *cli.Context) error {

			// Enable debug display if the flag is passed
			if c.Bool("debug") {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			} else {
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}

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
