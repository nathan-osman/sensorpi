package console

import (
	"fmt"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// Console provides a simple way to output values to STDOUT.
type Console struct{}

type outputParams struct {
	Label string `yaml:"label"`
}

func init() {
	plugin.Register("console", func(node *yaml.Node) (plugin.Plugin, error) {
		return &Console{}, nil
	})
}

func (c *Console) IsInput() bool {
	return false
}

func (c *Console) IsOutput() bool {
	return true
}

func (c *Console) Read(*yaml.Node) (float64, error) {
	return 0, nil
}

func (c *Console) Write(v float64, node *yaml.Node) error {
	params := &outputParams{}
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

func (c *Console) Close() {}
