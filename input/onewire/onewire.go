package onewire

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

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

func (o *OneWire) IsInput() bool {
	return true
}

func (o *OneWire) IsOutput() bool {
	return false
}

func (o *OneWire) Read(node *yaml.Node) (float64, error) {
	params := &inputParams{}
	if err := node.Decode(params); err != nil {
		return 0, err
	}
	f, err := os.Open(fmt.Sprintf("/sys/bus/w1/devices/%s/temperature", params.Device))
	if err != nil {
		return 0, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return 0, err
	}
	return float64(v) / 1000, nil
}

func (o *OneWire) Write(float64, *yaml.Node) error {
	return nil
}

func (o *OneWire) Close() {}
