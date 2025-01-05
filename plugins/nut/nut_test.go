package nut

import (
	"testing"

	"github.com/nathan-osman/sensorpi/plugin"
)

func TestPlugin(t *testing.T) {
	if !plugin.IsInputPlugin(&Nut{}) {
		t.Fatal("Nut does not correctly implement InputPlugin")
	}
}
