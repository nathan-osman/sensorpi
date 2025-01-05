package command

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsOutputPlugin(&Command{}) {
		t.Fatal("Command does not correctly implement OutputPlugin")
	}
}
