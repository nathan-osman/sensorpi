package onewire

import (
	"fmt"
	"io/ioutil"
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
	plugin.Register("onewire", func(node *yaml.Node) (any, error) {
		return &OneWire{}, nil
	})
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
	v, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	if err != nil {
		return 0, err
	}
	return float64(v) / 1000, nil
}
