//go:build !windows

package gpio

import (
	"context"
	"time"

	"github.com/nathan-osman/sensorpi/plugin"
	"github.com/stianeikeland/go-rpio/v4"
	"gopkg.in/yaml.v3"
)

type Gpio struct{}

type outputParams struct {
	Pin uint8 `yaml:"pin"`
}

type triggerParams struct {
	Pin          uint8  `yaml:"pin"`
	Invert       bool   `yaml:"invert"`
	PollInterval string `yaml:"poll_interval"`
}

type triggerData struct {
	Pin      uint8
	Invert   bool
	Duration time.Duration
}

func init() {
	plugin.Register("gpio", func(node *yaml.Node) (any, error) {
		if err := rpio.Open(); err != nil {
			return nil, err
		}
		return &Gpio{}, nil
	})
}

func (g *Gpio) WriteInit(node *yaml.Node) (any, error) {
	params := &outputParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	rpio.Pin(params.Pin).Output()
	return params, nil
}

func (g *Gpio) Write(data any, v float64) error {
	var (
		params = data.(*outputParams)
		state  = rpio.High
	)
	if v == 0 {
		state = rpio.Low
	}
	rpio.Pin(params.Pin).Write(state)
	return nil
}

func (g *Gpio) WatchInit(node *yaml.Node) (any, error) {
	params := &triggerParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	rpio.Pin(params.Pin).Detect(rpio.AnyEdge)
	var duration time.Duration
	if params.PollInterval != "" {
		d, err := time.ParseDuration(params.PollInterval)
		if err != nil {
			return nil, err
		}
		duration = d
	}
	if duration == 0 {
		duration = 100 * time.Millisecond
	}
	return &triggerData{
		Pin:      params.Pin,
		Invert:   params.Invert,
		Duration: duration,
	}, nil
}

func (g *Gpio) Watch(data any, ctx context.Context) (float64, error) {
	d := data.(*triggerData)

	// Pause for 100ms, to avoid phantom triggers from contact bounce
	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		return 0, context.Canceled
	}

	// TODO: use proper edge detection here instead of polling and then
	// reading the current value

	// Poll for rise / fall
	for {
		select {
		case <-time.After(d.Duration):
			if rpio.Pin(d.Pin).EdgeDetected() {
				var (
					isHigh = rpio.Pin(d.Pin).Read() == rpio.High
					v      float64
				)
				if d.Invert {
					isHigh = !isHigh
				}
				if isHigh {
					v = 1
				}
				return v, nil
			}
		case <-ctx.Done():
			return 0, context.Canceled
		}
	}
}

func (g *Gpio) Close() {
	rpio.Close()
}
