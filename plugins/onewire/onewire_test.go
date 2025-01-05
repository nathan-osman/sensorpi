package onewire

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsInputPlugin(&OneWire{}) {
		t.Fatal("OneWire does not correctly implement InputPlugin")
	}
}
