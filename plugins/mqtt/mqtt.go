package mqtt

import (
	"context"
	"fmt"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nathan-osman/sensorpi/plugin"
	"github.com/rs/zerolog/log"
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

type triggerParams struct {
	Topic string `yaml:"topic"`
	Qos   uint8  `yaml:"qos"`
}

type triggerData struct {
	Topic     string
	FloatChan <-chan float64
}

func init() {
	plugin.Register("mqtt", func(node *yaml.Node) (plugin.Plugin, error) {
		params := &pluginParams{}
		if err := node.Decode(params); err != nil {
			return nil, err
		}
		c := mqtt.NewClient(
			mqtt.NewClientOptions().
				AddBroker(fmt.Sprintf("tcp://%s", params.Addr)).
				SetClientID("sensorpi").
				SetResumeSubs(true).
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

func (m *Mqtt) WatchInit(node *yaml.Node) (any, error) {
	params := &triggerParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	fChan := make(chan float64)
	if t := m.client.Subscribe(
		params.Topic,
		params.Qos,
		func(client mqtt.Client, msg mqtt.Message) {
			v, err := strconv.ParseFloat(string(msg.Payload()), 64)
			if err != nil {
				log.Warn().Msgf("mqtt: %s", err.Error())
				return
			}
			fChan <- v
		},
	); t.Wait() && t.Error() != nil {
		return nil, t.Error()
	}
	return &triggerData{
		Topic:     params.Topic,
		FloatChan: fChan,
	}, nil
}

func (m *Mqtt) Watch(data any, ctx context.Context) (float64, error) {
	d := data.(*triggerData)
	select {
	case v := <-d.FloatChan:
		return v, nil
	case <-ctx.Done():
		return 0, context.Canceled
	}
}

func (m *Mqtt) WatchClose(data any) {
	d := data.(*triggerData)
	if t := m.client.Unsubscribe(d.Topic); t.Wait() && t.Error() != nil {
		log.Warn().Msgf("mqtt: %s", t.Error())
	}
}

func (m *Mqtt) Close() {
	m.client.Disconnect(1000)
}
