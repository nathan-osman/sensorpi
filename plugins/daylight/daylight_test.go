package daylight

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsInputPlugin(&Daylight{}) {
		t.Fatal("Daylight does not correctly implement InputPlugin")
	}
	if !plugin.IsTriggerPlugin(&Daylight{}) {
		t.Fatal("Daylight does not correctly implement TriggerPlugin")
	}
}
