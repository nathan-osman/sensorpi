package homeassistant

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsOutputPlugin(&HomeAssistant{}) {
		t.Fatal("HomeAssistant does not correctly implement OutputPlugin")
	}
	if !plugin.IsTriggerPlugin(&HomeAssistant{}) {
		t.Fatal("HomeAssistant does not correctly implement TriggerPlugin")
	}
}
