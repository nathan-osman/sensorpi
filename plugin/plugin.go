package plugin

import (
	"context"

	"gopkg.in/yaml.v3"
)

// InputPlugin represents a plugin that reads from a sensor.
type InputPlugin interface {

	// ReadInit initializes an instance of the plugin.
	ReadInit(*yaml.Node) (any, error)

	// Read collects the value for the provided input.
	Read(any) (float64, error)
}

// OutputPlugin represents a plugin that does something with data.
type OutputPlugin interface {

	// WriteInit initializes an instance of the plugin.
	WriteInit(*yaml.Node) (any, error)

	// Write processes the provided data.
	Write(any, float64) error
}

// TriggerPlugin represents a plugin that notifies when an event occurs.
type TriggerPlugin interface {

	// WatchInit initializes an instance of the plugin.
	WatchInit(*yaml.Node) (any, error)

	// Watch should wait until triggered or the context is cancelled. If no
	// error occurred, a float64 should be returned. If the context was
	// cancelled, context.Canceled should be returned.
	Watch(any, context.Context) (float64, error)
}
