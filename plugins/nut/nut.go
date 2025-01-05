package nut

import (
	"strconv"
	"time"

	"github.com/nathan-osman/nutclient/v3"
	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// Nut reads data from a NUT server.
type Nut struct {
	client *nutclient.Client
	name   string
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
				Addr:              params.Addr,
				KeepAliveInterval: 30 * time.Second,
			}),
			name: params.Name,
		}, nil
	})
}

func (n *Nut) ReadInit(node *yaml.Node) (any, error) {
	params := &inputParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	return params, nil
}

func (n *Nut) Read(data any) (float64, error) {
	params := data.(*inputParams)
	v, err := n.client.Get("VAR", n.name, params.Key)
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}

func (n *Nut) ReadClose(any) {}
