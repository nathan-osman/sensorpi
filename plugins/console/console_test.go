package console

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsOutputPlugin(&Console{}) {
		t.Fatal("Console does not correctly implement OutputPlugin")
	}
}
