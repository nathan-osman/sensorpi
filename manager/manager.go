package manager

import (
	"os"
	"time"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

type managerPluginAndParams struct {
	Plugin     plugin.Plugin
	Parameters *yaml.Node
}

type managerTask struct {
	Interval time.Duration
	NextRun  time.Time
	Input    *managerPluginAndParams
	Outputs  []*managerPluginAndParams
}

// Manager parses a configuration file and initializes inputs and outputs
// accordingly.
type Manager struct {
	plugins    map[string]plugin.Plugin
	tasks      []*managerTask
	closeChan  chan any
	closedChan chan any
}

type configPlugin struct {
	Plugin     string     `yaml:"plugin"`
	Parameters *yaml.Node `yaml:"parameters"`
}

type configConnection struct {
	Input    *configPlugin   `yaml:"input"`
	Outputs  []*configPlugin `yaml:"outputs"`
	Interval time.Duration   `yaml:"interval"`
}

type configRoot struct {
	Plugins     map[string]*yaml.Node `yaml:"plugins"`
	Connections []*configConnection   `yaml:"connections"`
}

func (m *Manager) getPlugin(name string, node *yaml.Node) (plugin.Plugin, error) {
	if p := m.plugins[name]; p != nil {
		return p, nil
	}
	p, err := plugin.Create(name, node)
	if err != nil {
		return nil, err
	}
	m.plugins[name] = p
	return p, nil
}

func (m *Manager) doTask(t *managerTask) error {
	v, err := t.Input.Plugin.Read(t.Input.Parameters)
	if err != nil {
		return err
	}
	for _, o := range t.Outputs {
		if err := o.Plugin.Write(v, o.Parameters); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) run() {
	defer close(m.closedChan)
	for {
		var (
			now      = time.Now()
			nextTask time.Duration
		)
		for _, t := range m.tasks {
			if !t.NextRun.After(now) {
				if err := m.doTask(t); err != nil {
					// TODO: log error
				}
			} else {
				n := t.NextRun.Sub(now)
				if nextTask == 0 || n < nextTask {
					nextTask = n
				}
			}
		}
		select {
		case <-time.After(nextTask):
		case <-m.closeChan:
			return
		}
	}
}

// New creates a new Manager instance and initializes it using the provided
// configuration file.
func New(filename string) (*Manager, error) {

	// Open the config file for reading
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Parse the root of the config file
	root := &configRoot{}
	if err := yaml.NewDecoder(f).Decode(root); err != nil {
		return nil, err
	}

	var (
		now = time.Now()
		m   = &Manager{
			plugins:    make(map[string]plugin.Plugin),
			closeChan:  make(chan any),
			closedChan: make(chan any),
		}
	)

	// Initialize any plugins that are explicitly specified
	for name, params := range root.Plugins {
		_, err := m.getPlugin(name, params)
		if err != nil {
			return nil, err
		}
	}

	// Enumerate the connections and create tasks for each of them
	for _, c := range root.Connections {
		p, err := m.getPlugin(c.Input.Plugin, nil)
		if err != nil {
			return nil, err
		}
		outputPlugins := []*managerPluginAndParams{}
		for _, output := range c.Outputs {
			p, err := m.getPlugin(output.Plugin, nil)
			if err != nil {
				return nil, err
			}
			outputPlugins = append(outputPlugins, &managerPluginAndParams{
				Plugin:     p,
				Parameters: output.Parameters,
			})
		}
		m.tasks = append(m.tasks, &managerTask{
			Interval: c.Interval,
			NextRun:  now,
			Input: &managerPluginAndParams{
				Plugin:     p,
				Parameters: c.Input.Parameters,
			},
			Outputs: outputPlugins,
		})
	}

	// Start the goroutine for processing the inputs
	go m.run()

	return m, nil
}

// Close shuts down the manager
func (m *Manager) Close() {
	close(m.closeChan)
	<-m.closedChan
}
