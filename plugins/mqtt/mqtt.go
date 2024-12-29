package mqtt

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// Mqtt maintains a connection to an MQTT server
type Mqtt struct {
	client mqtt.Client
}

type outputConfig struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type params struct {
	Topic  string `yaml:"topic"`
	Qos    uint8  `yaml:"qos"`
	Retain bool   `yaml:"retain"`
}

func init() {
	plugin.Register("mqtt", func(node *yaml.Node) (any, error) {
		cfg := &outputConfig{}
		if err := node.Decode(cfg); err != nil {
			return nil, err
		}
		c := mqtt.NewClient(
			mqtt.NewClientOptions().
				AddBroker(fmt.Sprintf("tcp://%s", cfg.Addr)).
				SetClientID("sensorpi").
				SetKeepAlive(30 * time.Second).
				SetPassword(cfg.Password).
				SetUsername(cfg.Username),
		)
		if t := c.Connect(); t.Wait() && t.Error() != nil {
			return nil, t.Error()
		}
		m := &Mqtt{
			client: c,
		}
		return m, nil
	})
}

func (m *Mqtt) Write(v float64, node *yaml.Node) error {
	params := &params{
		Qos:    1,
		Retain: true,
	}
	if err := node.Decode(params); err != nil {
		return err
	}
	if t := m.client.Publish(
		params.Topic,
		params.Qos,
		params.Retain,
		fmt.Sprintf("%f", v),
	); t.Wait() && t.Error() != nil {
		return t.Error()
	}
	return nil
}

func (m *Mqtt) Close() {
	m.client.Disconnect(1000)
}
