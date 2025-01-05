package mqtt

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsOutputPlugin(&Mqtt{}) {
		t.Fatal("Mqtt does not correctly implement OutputPlugin")
	}
}
