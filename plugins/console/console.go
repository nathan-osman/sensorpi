package console

import (
	"fmt"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// Console provides a simple way to output values to STDOUT.
type Console struct{}

type params struct {
	Label string `yaml:"label"`
}

func init() {
	plugin.Register("console", func(node *yaml.Node) (any, error) {
		return &Console{}, nil
	})
}

func printValue(v float64, node *yaml.Node) error {
	params := &params{}
	if err := node.Decode(params); err != nil {
		return err
	}
	label := params.Label
	if label == "" {
		label = "Value"
	}
	fmt.Printf("%s: %f\n", label, v)
	return nil
}

func (c *Console) Write(v float64, node *yaml.Node) error {
	return printValue(v, node)
}

func (c *Console) Run(v float64, node *yaml.Node) error {
	return printValue(v, node)
}
