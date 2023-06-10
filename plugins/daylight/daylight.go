package daylight

import (
	"context"
	"time"

	"github.com/nathan-osman/go-sunrise"
	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// Daylight triggers on sunrise and sunset.
type Daylight struct{}

type triggerParams struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
}

func init() {
	plugin.Register("daylight", func(node *yaml.Node) (any, error) {
		return &Daylight{}, nil
	})
}

func (d *Daylight) Watch(ctx context.Context, node *yaml.Node) (float64, error) {
	params := &triggerParams{}
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
