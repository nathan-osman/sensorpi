package nut

import (
	"errors"
	"strconv"
	"time"

	"github.com/nathan-osman/nutclient/v2"
	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

var errKeyNotFound = errors.New("key not found")

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
		return &Nut{
			client: nutclient.New(&nutclient.Config{
				Addr: params.Addr,
				Name: params.Name,
			}),
		}, nil
	})
}

func (n *Nut) Read(node *yaml.Node) (float64, error) {
	params := &inputParams{}
	if err := node.Decode(params); err != nil {
		return 0, err
	}
	l, err := n.client.Status()
	if err != nil {
		return 0, err
	}
	v, ok := l[params.Key]
	if !ok {
		return 0, errKeyNotFound
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}
