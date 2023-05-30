package plugin

import (
	"gopkg.in/yaml.v3"
)

// InputPlugin represents a plugin that reads from a sensor.
type InputPlugin interface {

	// Read collects the value for the provided input.
	Read(*yaml.Node) (float64, error)
}

// OutputPlugin represents a plugin that does something with data.
type OutputPlugin interface {

	// Write processes the provided data.
	Write(float64, *yaml.Node) error
}
