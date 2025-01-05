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

type triggerData struct {
	Duration time.Duration
}

func init() {
	plugin.Register("timer", func(node *yaml.Node) (any, error) {
		return &Timer{}, nil
	})
}

func (t *Timer) WatchInit(node *yaml.Node) (any, error) {
	params := &triggerParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	d, err := time.ParseDuration(params.Interval)
	if err != nil {
		return nil, err
	}
	return &triggerData{
		Duration: d,
	}, nil
}

func (t *Timer) Watch(data any, ctx context.Context) (float64, error) {
	d := data.(*triggerData)
	select {
	case <-time.After(d.Duration):
		return 0, nil
	case <-ctx.Done():
		return 0, context.Canceled
	}
}

func (t *Timer) WatchClose(any) {}
