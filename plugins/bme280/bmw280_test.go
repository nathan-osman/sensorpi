package bme280

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsInputPlugin(&BME280{}) {
		t.Fatal("BME280 does not correctly implement InputPlugin")
	}
}
