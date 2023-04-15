package plugin

import (
	"gopkg.in/yaml.v3"
)

// Plugin represents a type that provides sensor values or processes them in
// some way. For example, a plugin might read from a sensor or write to a log
// file.
type Plugin interface {

	// IsInput indicates that this plugin can be used as an input.
	IsInput() bool

	// IsOutput indicates that this plugin can be used as an output.
	IsOutput() bool

	// Read collects the value for the provided input.
	Read(*yaml.Node) (float64, error)

	// Write processes the provided data.
	Write(float64, *yaml.Node) error

	// Close shuts down the plugin.
	Close()
}
