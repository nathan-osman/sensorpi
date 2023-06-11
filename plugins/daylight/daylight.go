package daylight

import (
	"context"
	"time"

	"github.com/nathan-osman/go-sunrise"
	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// Daylight, when used as an input returns 1.0 for daylight, and when used as a
// trigger, triggers on sunrise and sunset.
type Daylight struct{}

type params struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
}

func init() {
	plugin.Register("daylight", func(node *yaml.Node) (any, error) {
		return &Daylight{}, nil
	})
}

func (d *Daylight) Read(node *yaml.Node) (float64, error) {
	params := &params{}
	if err := node.Decode(params); err != nil {
		return 0, err
	}
	var (
		t      = time.Now()
		sr, ss = sunrise.SunriseSunset(
			params.Latitude,
			params.Longitude,
			t.Year(),
			t.Month(),
			t.Day(),
		)
	)
	if t.After(sr) && t.Before(ss) {
		return 1, nil
	} else {
		return 0, nil
	}
}

func (d *Daylight) Watch(ctx context.Context, node *yaml.Node) (float64, error) {
	params := &params{}
	if err := node.Decode(params); err != nil {
		return 0, err
	}
	var (
		tNow   = time.Now()
		tNext  time.Time
		sr, ss = sunrise.SunriseSunset(
			params.Latitude,
			params.Longitude,
			tNow.Year(),
			tNow.Month(),
			tNow.Day(),
		)
	)
	switch {
	case sr.After(tNow):
		tNext = sr
	case ss.After(tNow):
		tNext = ss
	default:
		sr, _ = sunrise.SunriseSunset(
			params.Latitude,
			params.Longitude,
			tNow.Year(),
			tNow.Month(),
			tNow.Day()+1,
		)
		tNext = sr
	}
	select {
	case <-time.After(tNext.Sub(tNow)):
		return 0, nil
	case <-ctx.Done():
		return 0, context.Canceled
	}
}
