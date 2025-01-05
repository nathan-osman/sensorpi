package timer

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsTriggerPlugin(&Timer{}) {
		t.Fatal("Timer does not correctly implement TriggerPlugin")
	}
}
