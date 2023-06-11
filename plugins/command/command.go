package command

import (
	"errors"
	"os/exec"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// Command executes a command as an action.
type Command struct{}

type params struct {
	Name string   `yaml:"name"`
	Args []string `yaml:"arguments"`
}

func init() {
	plugin.Register("command", func(node *yaml.Node) (any, error) {
		return &Command{}, nil
	})
}

func (c *Command) Write(v float64, node *yaml.Node) error {
	params := &params{}
	if err := node.Decode(params); err != nil {
		return err
	}
	if err := exec.Command(params.Name, params.Args...).Run(); err != nil {
		e, ok := err.(*exec.ExitError)
		if ok {
			return errors.New(string(e.Stderr))
		}
		return err
	}
	return nil
}
