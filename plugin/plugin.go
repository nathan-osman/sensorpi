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

	// Watch should wait until triggered or the context is cancelled.
	Watch(context.Context, *yaml.Node) error
}

// ActionPlugin represents a plugin that responds to a trigger.
type ActionPlugin interface {

	// Run invokes the action.
	Run(*yaml.Node) error
}
