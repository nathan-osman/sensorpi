package manager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
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
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
	closeChan  chan any
	closedChan chan any
}

// TODO: there is a lot of duplication but composition requires exported types

type configRoot struct {
	Plugins map[string]yaml.Node `yaml:"plugins"`
	Inputs  []*struct {
		Plugin     string    `yaml:"plugin"`
		Parameters yaml.Node `yaml:"parameters"`
		Outputs    []*struct {
			Plugin     string    `yaml:"plugin"`
			Parameters yaml.Node `yaml:"parameters"`
		} `yaml:"outputs"`
		Interval time.Duration `yaml:"interval"`
	} `yaml:"inputs"`
	Triggers []*struct {
		Plugin     string    `yaml:"plugin"`
		Parameters yaml.Node `yaml:"parameters"`
		Outputs    []*struct {
			Plugin     string    `yaml:"plugin"`
			Parameters yaml.Node `yaml:"parameters"`
		} `yaml:"outputs"`
	} `yaml:"triggers"`
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
		now             = time.Now()
		ctx, cancelFunc = context.WithCancel(context.Background())
		m               = &Manager{
			plugins:    make(map[string]any),
			cancelFunc: cancelFunc,
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

	// Enumerate the inputs and create tasks for each of them
	for _, i := range root.Inputs {
		v, err := m.getPlugin(i.Plugin, nil)
		if err != nil {
			return nil, err
		}
		p, ok := v.(plugin.InputPlugin)
		if !ok {
			return nil, fmt.Errorf("%s is not an input plugin", i.Plugin)
		}
		if i.Interval == 0 {
			return nil, errors.New("interval cannot be zero")
		}
		outputPlugins := []*managerOutputPluginAndParams{}
		for _, output := range i.Outputs {
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
			Interval: i.Interval,
			NextRun:  now,
			Input: &managerInputPluginAndParams{
				Name:       i.Plugin,
				Plugin:     p,
				Parameters: &i.Parameters,
			},
			Outputs: outputPlugins,
		})
	}

	// Enumerate the triggers
	for _, t := range root.Triggers {
		v, err := m.getPlugin(t.Plugin, nil)
		if err != nil {
			return nil, err
		}
		p, ok := v.(plugin.TriggerPlugin)
		if !ok {
			return nil, fmt.Errorf("%s is not a trigger plugin", t.Plugin)
		}
		actions := []*managerOutputPluginAndParams{}
		for _, action := range t.Outputs {
			v, err := m.getPlugin(action.Plugin, nil)
			if err != nil {
				return nil, err
			}
			p, ok := v.(plugin.OutputPlugin)
			if !ok {
				return nil, fmt.Errorf("%s is not an output plugin", action.Plugin)
			}
			actions = append(actions, &managerOutputPluginAndParams{
				Name:       action.Plugin,
				Plugin:     p,
				Parameters: &action.Parameters,
			})
		}
		m.wg.Add(1)
		go func(name string, node *yaml.Node) {
			defer m.wg.Done()
			for {
				v, err := p.Watch(ctx, node)
				if err != nil {
					if err == context.Canceled {
						return
					} else {
						log.Error().Msg(err.Error())
					}
				}
				log.Debug().Msgf("triggered %f from %s", v, name)
				for _, a := range actions {
					if err := a.Plugin.Write(v, a.Parameters); err != nil {
						log.Error().Msg(err.Error())
					}
				}
			}
		}(t.Plugin, &t.Parameters)
	}

	// Abort if there are no tasks
	if len(m.tasks) == 0 {
		return nil, errors.New("no tasks were created; aborting")
	}

	// Start the goroutine for processing the inputs
	go m.run()

	return m, nil
}

// Close shuts down the manager.
func (m *Manager) Close() {

	// Shut down the task goroutine
	close(m.closeChan)

	// Shut down all of the goroutine monitoring triggers
	m.cancelFunc()

	// Wait for goroutines to finish
	<-m.closedChan
	m.wg.Wait()
}
