//go:build !windows

package gpio

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

type Gpio struct{}

type outputParams struct {
	Pin int `yaml:"pin"`
}

type outputData struct {
	pin gpio.PinIO
}

type triggerParams struct {
	Pin              int    `yaml:"pin"`
	Invert           bool   `yaml:"invert"`
	PullUpDown       string `yaml:"pull_up_down"`
	DebounceInterval string `yaml:"debounce_interval"`
}

type triggerData struct {
	watcher          *gpioWatcher
	invert           bool
	lastLevel        gpio.Level
	lastLevelTime    time.Time
	debounceDuration time.Duration
}

func init() {
	plugin.Register("gpio", func(node *yaml.Node) (plugin.Plugin, error) {
		_, err := host.Init()
		if err != nil {
			return nil, err
		}
		return &Gpio{}, nil
	})
}

func pinByName(pin int) (gpio.PinIO, error) {
	p := gpioreg.ByName(strconv.Itoa(pin))
	if p == nil {
		return nil, fmt.Errorf("GPIO pin %d does not exist", pin)
	}
	return p, nil
}

func parsePullUpDown(v string) (gpio.Pull, error) {
	switch v {
	case "":
		return gpio.Float, nil
	case "up":
		return gpio.PullUp, nil
	case "down":
		return gpio.PullDown, nil
	default:
		return 0, fmt.Errorf("invalid value %s for built-in resistor", v)
	}
}

func (g *Gpio) WriteInit(node *yaml.Node) (any, error) {
	params := &outputParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	p, err := pinByName(params.Pin)
	if err != nil {
		return nil, err
	}
	return &outputData{
		pin: p,
	}, nil
}

func (g *Gpio) Write(data any, v float64) error {
	var (
		d = data.(*outputData)
		l = gpio.High
	)
	if v == 0 {
		l = gpio.Low
	}
	return d.pin.Out(l)
}

func (g *Gpio) WriteClose(any) {}

func (g *Gpio) WatchInit(node *yaml.Node) (any, error) {
	params := &triggerParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	p, err := pinByName(params.Pin)
	if err != nil {
		return nil, err
	}
	pull, err := parsePullUpDown(params.PullUpDown)
	if err != nil {
		return nil, err
	}
	if err := p.In(pull, gpio.BothEdges); err != nil {
		return nil, err
	}
	lastLevel := gpio.Low
	if params.Invert {
		lastLevel = gpio.High
	}
	var debounceDuration time.Duration
	if params.DebounceInterval != "" {
		d, err := time.ParseDuration(params.DebounceInterval)
		if err != nil {
			return nil, err
		}
		debounceDuration = d
	}
	if debounceDuration == 0 {
		debounceDuration = 200 * time.Millisecond
	}
	return &triggerData{
		watcher:          newGpioWatcher(p),
		invert:           params.Invert,
		lastLevel:        lastLevel,
		debounceDuration: debounceDuration,
	}, nil
}

func (g *Gpio) Watch(data any, ctx context.Context) (float64, error) {
	d := data.(*triggerData)
	for {
		select {
		case <-d.watcher.edgeChan:

			var n = time.Now()

			// Don't do anything if the debounce interval hasn't elapsed
			if n.Before(d.lastLevelTime.Add(d.debounceDuration)) {
				continue
			}

			// The pin has changed value, so invert it
			d.lastLevel = !d.lastLevel
			d.lastLevelTime = n

			var (
				isHigh = d.lastLevel == gpio.High
				v      float64
			)
			if d.invert {
				isHigh = !isHigh
			}
			if isHigh {
				v = 1
			}

			// Return the new value
			return v, nil

		case <-ctx.Done():

			return 0, context.Canceled
		}
	}
}

func (g *Gpio) WatchClose(data any) {
	data.(*triggerData).watcher.Close()
}

func (g *Gpio) Close() {}
