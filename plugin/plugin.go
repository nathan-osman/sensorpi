package plugin

import (
	"context"

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

// TriggerPlugin represents a plugin that notifies when an event occurs.
type TriggerPlugin interface {

	// Watch should wait until triggered or the context is cancelled. If no
	// error occurred, a float64 should be returned. If the context was
	// cancelled, context.Canceled should be returned.
	Watch(context.Context, *yaml.Node) (float64, error)
}

// ActionPlugin represents a plugin that responds to a trigger or a specific
// condition when reading values from an InputPlugin.
type ActionPlugin interface {

	// Run invokes the action.
	Run(float64, *yaml.Node) error
}
