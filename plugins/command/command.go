package command

import (
	"errors"
	"os/exec"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// Command executes a command as an action.
type Command struct{}

type outputParams struct {
	Name string   `yaml:"name"`
	Args []string `yaml:"arguments"`
}

func init() {
	plugin.Register("command", func(node *yaml.Node) (any, error) {
		return &Command{}, nil
	})
}

func (c *Command) WriteInit(node *yaml.Node) (any, error) {
	params := &outputParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	return params, nil
}

func (c *Command) Write(data any, v float64) error {
	params := data.(*outputParams)
	if err := exec.Command(params.Name, params.Args...).Run(); err != nil {
		e, ok := err.(*exec.ExitError)
		if ok {
			s := string(e.Stderr)
			if len(s) != 0 {
				return errors.New(s)
			}
		}
		return err
	}
	return nil
}

func (c *Command) WriteClose(any) {}
