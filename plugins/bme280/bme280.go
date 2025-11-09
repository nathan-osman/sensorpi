package bme280

import (
	"errors"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

const (
	quantityTemperature = "temperature"
	quantityHumidity    = "humidity"
	quantityPressure    = "pressure"
)

// BME280 provides access to a BME280 sensor.
type BME280 struct {
	bus i2c.BusCloser
}

type inputParams struct {
	Address  uint16 `yaml:"address"`
	Quantity string `yaml:"quantity"`
}

type inputData struct {
	dev      *bmxx80.Dev
	quantity string
}

func init() {
	plugin.Register("bme280", func(node *yaml.Node) (plugin.Plugin, error) {
		_, err := host.Init()
		if err != nil {
			return nil, err
		}
		b, err := i2creg.Open("1")
		if err != nil {
			return nil, err
		}
		return &BME280{
			bus: b,
		}, nil
	})
}

func (b *BME280) ReadInit(node *yaml.Node) (any, error) {
	params := &inputParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	d, err := bmxx80.NewI2C(b.bus, 0x76, &bmxx80.DefaultOpts)
	if err != nil {
		return nil, err
	}
	return &inputData{
		dev:      d,
		quantity: params.Quantity,
	}, nil
}

func (b *BME280) Read(data any) (float64, error) {
	var (
		d   = data.(*inputData)
		env physic.Env
	)
	if err := d.dev.Sense(&env); err != nil {
		return 0, err
	}
	switch d.quantity {
	case quantityTemperature:
		return env.Temperature.Celsius(), nil
	case quantityHumidity:
		return float64(env.Humidity), nil
	case quantityPressure:
		return float64(env.Pressure), nil
	default:
		return 0, errors.New("invalid quantity specified")
	}
}

func (b *BME280) ReadClose(data any) {
	d := data.(*inputData)
	d.dev.Halt()
}

// Close shuts down the plugin.
func (b *BME280) Close() {
	b.bus.Close()
}
