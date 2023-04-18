package plugin

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// FactoryFn provides a method for initializing a plugin.
type FactoryFn func(*yaml.Node) (Plugin, error)

var (
	pluginMap map[string]FactoryFn = make(map[string]FactoryFn)
)

// Register registers a plugin in the global plugin map.
func Register(name string, factoryFn FactoryFn) {
	pluginMap[name] = factoryFn
}

// Create attempts to create a new plugin instance.
func Create(name string, node *yaml.Node) (Plugin, error) {
	f := pluginMap[name]
	if f == nil {
		return nil, fmt.Errorf("unknown plugin \"%s\"", name)
	}
	return f(node)
}
