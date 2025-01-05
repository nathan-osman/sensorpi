package moisture

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsInputPlugin(&Moisture{}) {
		t.Fatal("Moisture does not correctly implement InputPlugin")
	}
}
