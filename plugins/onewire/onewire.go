package onewire

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// OneWire reads data from a OneWire sensor.
type OneWire struct{}

type inputParams struct {
	Device string `yaml:"device"`
}

func init() {
	plugin.Register("onewire", func(node *yaml.Node) (plugin.Plugin, error) {
		return &OneWire{}, nil
	})
}

func (o *OneWire) ReadInit(node *yaml.Node) (any, error) {
	params := &inputParams{}
	if err := node.Decode(params); err != nil {
		return 0, err
	}
	return params, nil
}

func (o *OneWire) Read(data any) (float64, error) {
	params := data.(*inputParams)
	f, err := os.Open(fmt.Sprintf("/sys/bus/w1/devices/%s/temperature", params.Device))
	if err != nil {
		return 0, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	if err != nil {
		return 0, err
	}
	return float64(v) / 1000, nil
}

func (o *OneWire) ReadClose(any) {}

func (o *OneWire) Close() {}
