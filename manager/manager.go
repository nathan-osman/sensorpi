package manager

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/nathan-osman/sensorpi/plugin"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type managerInputPluginAndParams struct {
	Name       string
	Plugin     plugin.InputPlugin
	Parameters *yaml.Node
}

type managerOutputPluginAndParams struct {
	Name       string
	Plugin     plugin.OutputPlugin
	Parameters *yaml.Node
}

type managerTask struct {
	Interval time.Duration
	NextRun  time.Time
	Input    *managerInputPluginAndParams
	Outputs  []*managerOutputPluginAndParams
}

// Manager parses a configuration file and initializes inputs and outputs
// accordingly.
type Manager struct {
	plugins    map[string]any
	tasks      []*managerTask
	closeChan  chan any
	closedChan chan any
}

type configPlugin struct {
	Plugin     string    `yaml:"plugin"`
	Parameters yaml.Node `yaml:"parameters"`
}

type configConnection struct {
	Input    *configPlugin   `yaml:"input"`
	Outputs  []*configPlugin `yaml:"outputs"`
	Interval time.Duration   `yaml:"interval"`
}

type configTrigger struct {
	configPlugin
	Actions []*configPlugin `yaml:"actions"`
}

type configRoot struct {
	Plugins     map[string]yaml.Node `yaml:"plugins"`
	Connections []*configConnection  `yaml:"connections"`
	Triggers    []*configTrigger     `yaml:"triggers"`
}

func (m *Manager) getPlugin(name string, node *yaml.Node) (any, error) {
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
	log.Debug().Msgf("read %f from %s", v, t.Input.Name)
	for _, o := range t.Outputs {
		if err := o.Plugin.Write(v, o.Parameters); err != nil {
			log.Error().Msg(err.Error())
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
					log.Error().Msg(err.Error())
				}
				t.NextRun = t.NextRun.Add(t.Interval)
			}
			n := t.NextRun.Sub(now)
			if nextTask == 0 || n < nextTask {
				nextTask = n
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
			plugins:    make(map[string]any),
			closeChan:  make(chan any),
			closedChan: make(chan any),
		}
	)

	// Initialize any plugins that are explicitly specified
	for name, params := range root.Plugins {
		_, err := m.getPlugin(name, &params)
		if err != nil {
			return nil, err
		}
	}

	// Enumerate the connections and create tasks for each of them
	for _, c := range root.Connections {
		v, err := m.getPlugin(c.Input.Plugin, nil)
		if err != nil {
			return nil, err
		}
		p, ok := v.(plugin.InputPlugin)
		if !ok {
			return nil, fmt.Errorf("%s is not an input plugin", c.Input.Plugin)
		}
		outputPlugins := []*managerOutputPluginAndParams{}
		for _, output := range c.Outputs {
			v, err := m.getPlugin(output.Plugin, nil)
			if err != nil {
				return nil, err
			}
			p, ok := v.(plugin.OutputPlugin)
			if !ok {
				return nil, fmt.Errorf("%s is not an output plugin", output.Plugin)
			}
			outputPlugins = append(outputPlugins, &managerOutputPluginAndParams{
				Name:       output.Plugin,
				Plugin:     p,
				Parameters: &output.Parameters,
			})
		}
		m.tasks = append(m.tasks, &managerTask{
			Interval: c.Interval,
			NextRun:  now,
			Input: &managerInputPluginAndParams{
				Name:       c.Input.Plugin,
				Plugin:     p,
				Parameters: &c.Input.Parameters,
			},
			Outputs: outputPlugins,
		})
	}

	// Enumerate the triggers
	for _, t := range root.Triggers {
		//...
	}

	// Abort if there are no tasks
	if len(m.tasks) == 0 {
		return nil, errors.New("no tasks were created; aborting")
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
