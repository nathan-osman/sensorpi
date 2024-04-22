package nut

import (
	"errors"
	"strconv"
	"time"

	"github.com/nathan-osman/nutclient"
	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

var (
	errNoStatusAvailable = errors.New("no status available")
	errKeyNotFound       = errors.New("key not found")
)

// Nut reads data from a NUT server.
type Nut struct {
	client *nutclient.Client
}

type pluginParams struct {
	Addr         string        `yaml:"addr"`
	Name         string        `yaml:"name"`
	PollInterval time.Duration `yaml:"poll_interval"`
}

type inputParams struct {
	Key string `yaml:"key"`
}

func init() {
	plugin.Register("nut", func(node *yaml.Node) (any, error) {
		params := &pluginParams{}
		if err := node.Decode(params); err != nil {
			return 0, err
		}
		return nutclient.New(&nutclient.Config{
			Addr:         params.Addr,
			Name:         params.Name,
			PollInterval: params.PollInterval,
		}), nil
	})
}

func (n *Nut) Read(node *yaml.Node) (float64, error) {
	params := &inputParams{}
	if err := node.Decode(params); err != nil {
		return 0, err
	}
	s := n.client.Status()
	if s == nil {
		return 0, errNoStatusAvailable
	}
	v, ok := s[params.Key]
	if !ok {
		return 0, errKeyNotFound
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}
