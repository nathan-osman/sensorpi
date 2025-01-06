package plugin

import (
	"context"

	"gopkg.in/yaml.v3"
)

// Plugin must be implemented by every plugin type.
type Plugin interface {

	// Close frees any resources used by the plugin.
	Close()
}

// InputPlugin represents a plugin that reads from a sensor.
type InputPlugin interface {

	// ReadInit initializes an instance of the plugin.
	ReadInit(*yaml.Node) (any, error)

	// Read collects the value for the provided input.
	Read(any) (float64, error)

	// ReadClose performs any cleanup from ReadInit.
	ReadClose(any)
}

// OutputPlugin represents a plugin that does something with data.
type OutputPlugin interface {

	// WriteInit initializes an instance of the plugin.
	WriteInit(*yaml.Node) (any, error)

	// Write processes the provided data.
	Write(any, float64) error

	// WriteClose performs any cleanup from WriteInit.
	WriteClose(any)
}

// TriggerPlugin represents a plugin that notifies when an event occurs.
type TriggerPlugin interface {

	// WatchInit initializes an instance of the plugin.
	WatchInit(*yaml.Node) (any, error)

	// Watch should wait until triggered or the context is cancelled. If no
	// error occurred, a float64 should be returned. If the context was
	// cancelled, context.Canceled should be returned.
	Watch(any, context.Context) (float64, error)

	// WatchClose performs any cleanup from WatchInit.
	WatchClose(any)
}
