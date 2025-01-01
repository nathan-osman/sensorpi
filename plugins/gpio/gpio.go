package gpio

import (
	"context"
	"time"

	"github.com/nathan-osman/sensorpi/plugin"
	"github.com/stianeikeland/go-rpio/v4"
	"gopkg.in/yaml.v3"
)

type Gpio struct{}

type writeParams struct {
	Pin uint8 `yaml:"pin"`
}

type triggerParams struct {
	Pin           uint8         `yaml:"pin"`
	PollInterval  time.Duration `yaml:"poll_interval"`
	TriggerOnRise bool          `yaml:"trigger_on_rise"`
}

func init() {
	plugin.Register("gpio", func(node *yaml.Node) (any, error) {
		if err := rpio.Open(); err != nil {
			return nil, err
		}
		return &Gpio{}, nil
	})
}

func (g *Gpio) Write(v float64, node *yaml.Node) error {
	params := &writeParams{}
	if err := node.Decode(params); err != nil {
		return err
	}
	rpio.Pin(params.Pin).Output()
	state := rpio.High
	if v == 0 {
		state = rpio.Low
	}
	rpio.Pin(params.Pin).Write(state)
	return nil
}

func (g *Gpio) Watch(ctx context.Context, node *yaml.Node) (float64, error) {
	params := &triggerParams{}
	if err := node.Decode(params); err != nil {
		return 0, err
	}

	// Pause for 100ms, to avoid phantom triggers from contact bounce
	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		return 0, context.Canceled
	}

	// Enable edge detection
	var edge = rpio.FallEdge
	if params.TriggerOnRise {
		edge = rpio.RiseEdge
	}
	rpio.Pin(params.Pin).Detect(edge)

	// Poll for rise / fall
	pollInterval := params.PollInterval
	if pollInterval == 0 {
		pollInterval = 100 * time.Millisecond
	}
	for {
		select {
		case <-time.After(pollInterval):
			if rpio.Pin(params.Pin).EdgeDetected() {
				return 0, nil
			}
		case <-ctx.Done():
			return 0, context.Canceled
		}
	}
}

func (g *Gpio) Close() {
	rpio.Close()
}
