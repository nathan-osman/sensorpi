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
	plugin.Register("console", func(node *yaml.Node) (any, error) {
		return &Console{}, nil
	})
}

func (c *Console) WriteInit(node *yaml.Node) (any, error) {
	params := &outputParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	return params, nil
}

func (c *Console) Write(data any, v float64) error {
	var (
		params = data.(*outputParams)
		label  = params.Label
	)
	if label == "" {
		label = "Value"
	}
	fmt.Printf("%s: %f\n", label, v)
	return nil
}
