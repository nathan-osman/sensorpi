package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/urfave/cli/v2"
)

const systemdUnitFile = `[Unit]
Description=sensorpi
Wants=network-online.target
After=network-online.target

[Service]
ExecStart={{.path}} --config {{.config_path}}

[Install]
WantedBy=multi-user.target
`

const configFile = `# TODO: use this file to configure the application

plugins:
  # Plugin-specific configuration here
connections:
  # Connect plugins here
`

var installCommand = &cli.Command{
	Name:  "install",
	Usage: "install the application as a local service",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Value: "/etc/sensorpi/config.yaml",
			Usage: "filename of configuration file",
		},
	},
	Action: install,
}

func writeTemplate(filename, content string, data any) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	t, err := template.New("").Parse(content)
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, data)
}

func install(c *cli.Context) error {

	// Write the configuration file
	if err := writeTemplate(
		c.String("config"),
		configFile,
		nil,
	); err != nil {
		return err
	}

	// Determine the full path to the executable
	p, err := os.Executable()
	if err != nil {
		return err
	}

	// Write the unit file
	if err := writeTemplate(
		"/lib/systemd/system/sensorpi.service",
		systemdUnitFile,
		map[string]interface{}{
			"path":        p,
			"config_path": c.String("config"),
		},
	); err != nil {
		return err
	}

	fmt.Println("Service installed!")
	fmt.Println("")
	fmt.Println("To enable the service and start it, run:")
	fmt.Println("  systemctl enable sensorpi")
	fmt.Println("  systemctl start sensorpi")

	return nil
}
