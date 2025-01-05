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

type pluginParams struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type outputParams struct {
	Topic  string `yaml:"topic"`
	Qos    uint8  `yaml:"qos"`
	Retain bool   `yaml:"retain"`
}

func init() {
	plugin.Register("mqtt", func(node *yaml.Node) (any, error) {
		params := &pluginParams{}
		if err := node.Decode(params); err != nil {
			return nil, err
		}
		c := mqtt.NewClient(
			mqtt.NewClientOptions().
				AddBroker(fmt.Sprintf("tcp://%s", params.Addr)).
				SetClientID("sensorpi").
				SetKeepAlive(30 * time.Second).
				SetPassword(params.Password).
				SetUsername(params.Username),
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

func (m *Mqtt) WriteInit(node *yaml.Node) (any, error) {
	params := &outputParams{
		Qos:    1,
		Retain: true,
	}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	return params, nil
}

func (m *Mqtt) Write(data any, v float64) error {
	params := data.(*outputParams)
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

func (m *Mqtt) WriteClose(any) {}

func (m *Mqtt) Close() {
	m.client.Disconnect(1000)
}
