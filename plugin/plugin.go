package plugin

import (
	"gopkg.in/yaml.v3"
)

// Plugin represents a type that provides sensor values or processes them in
// some way.
type Plugin interface {

	// IsInput indicates that this plugin can be used as an input.
	IsInput() bool

	// IsOutput indicates that this plugin can be used as an output.
	IsOutput() bool

	// Read collects the value for the specified input.
	Read(*yaml.Node) (float64, error)

	// Write processes data.
	Write(float64, *yaml.Node) error

	// Close shuts down the plugin.
	Close()
}
