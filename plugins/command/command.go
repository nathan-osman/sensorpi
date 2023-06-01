package command

import (
	"os/exec"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// Command executes a command as an action.
type Command struct{}

type actionParams struct {
	Name string   `yaml:"name"`
	Args []string `yaml:"arguments"`
}

func init() {
	plugin.Register("command", func(node *yaml.Node) (any, error) {
		return &Command{}, nil
	})
}

func (c *Command) Run(node *yaml.Node) error {
	params := &actionParams{}
	if err := node.Decode(params); err != nil {
		return err
	}
	return exec.Command(params.Name, params.Args...).Run()
}
