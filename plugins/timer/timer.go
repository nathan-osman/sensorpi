package timer

import (
	"context"
	"time"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// Timer triggers on the provided interval.
type Timer struct{}

type triggerParams struct {
	Interval string `yaml:"interval"`
}

func init() {
	plugin.Register("timer", func(node *yaml.Node) (any, error) {
		return &Timer{}, nil
	})
}

func (t *Timer) Watch(ctx context.Context, node *yaml.Node) (float64, error) {
	params := &triggerParams{}
	if err := node.Decode(params); err != nil {
		return 0, err
	}
	d, err := time.ParseDuration(params.Interval)
	if err != nil {
		return 0, err
	}
	select {
	case <-time.After(d):
		return 0, nil
	case <-ctx.Done():
		return 0, context.Canceled
	}
}
