package influxdb

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsOutputPlugin(&InfluxDB{}) {
		t.Fatal("InfluxDB does not correctly implement OutputPlugin")
	}
}
