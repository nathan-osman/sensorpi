package plugin

// Plugin represents a type that provides sensor values or processes them in
// some way.
type Plugin interface {

	// IsInput indicates that this plugin can be used as an input.
	IsInput() bool

	// IsOutput indicates that this plugin can be used as an output.
	IsOutput() bool

	// Close shuts down the plugin.
	Close()
}
