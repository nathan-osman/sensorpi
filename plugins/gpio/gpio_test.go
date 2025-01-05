//go:build !windows

package gpio

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsOutputPlugin(&Gpio{}) {
		t.Fatal("Gpio does not correctly implement OutputPlugin")
	}
	if !plugin.IsTriggerPlugin(&Gpio{}) {
		t.Fatal("Gpio does not correctly implement TriggerPlugin")
	}
}
