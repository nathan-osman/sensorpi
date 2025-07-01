package homeassistant

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nathan-osman/sensorpi/plugin"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	typeLight   = "light"
	typeSensor  = "sensor"
	typeTrigger = "trigger"
)

// HomeAssistant uses MQTT (with discovery) to interact with Home Assistant
type HomeAssistant struct {
	client      mqtt.Client
	nodeId      string
	actionTopic string
	device      map[string]any
}

type pluginParams struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	NodeId   string `yaml:"node_id"`
}

type outputTriggerParams struct {
	Type       string    `yaml:"type"`
	Parameters yaml.Node `yaml:"parameters"`
}

type outputParamsSensor struct {
	ID                        string `yaml:"id"`
	Name                      string `yaml:"name"`
	Class                     string `yaml:"class"`
	UnitOfMeasurement         string `yaml:"unit_of_measurement"`
	SuggestedDisplayPrecision string `yaml:"suggested_display_precision"`
}

type outputParamsTrigger struct {
	Type    string `yaml:"type"`
	Subtype string `yaml:"subtype"`
}

type outputData interface {
	Write(*HomeAssistant, float64) error
}

type outputDataSensor struct {
	topic string
}

type outputDataTrigger struct {
	subtype string
}

type triggerParamsLight struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type triggerData interface {
	Watch(*HomeAssistant, context.Context) (float64, error)
	Close(*HomeAssistant)
}

type triggerDataLight struct {
	topic   string
	cmdChan <-chan bool
}

func init() {
	plugin.Register("homeassistant", func(node *yaml.Node) (plugin.Plugin, error) {
		params := &pluginParams{}
		if err := node.Decode(params); err != nil {
			return nil, err
		}
		if params.NodeId == "" {
			h, err := os.Hostname()
			if err != nil {
				return nil, err
			}
			params.NodeId = h
		}
		c := mqtt.NewClient(
			mqtt.NewClientOptions().
				AddBroker(fmt.Sprintf("tcp://%s", params.Addr)).
				SetClientID(params.NodeId).
				SetResumeSubs(true).
				SetPassword(params.Password).
				SetUsername(params.Username),
		)
		if t := c.Connect(); t.Wait() && t.Error() != nil {
			return nil, t.Error()
		}
		h := &HomeAssistant{
			client: c,
			nodeId: params.NodeId,
			actionTopic: fmt.Sprintf(
				"sensorpi/%s/action",
				params.NodeId,
			),
			device: map[string]any{
				"identifiers": []string{
					fmt.Sprintf("sensorpi_%s", params.NodeId),
				},
				"name": params.NodeId,
			},
		}
		return h, nil
	})
}

func (h *HomeAssistant) WriteInit(node *yaml.Node) (any, error) {
	params := &outputTriggerParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	switch params.Type {
	case typeSensor:
		cParams := &outputParamsSensor{}
		if err := params.Parameters.Decode(cParams); err != nil {
			return nil, err
		}
		var (
			topic = fmt.Sprintf(
				"homeassistant/sensor/%s/%s/config",
				h.nodeId,
				cParams.ID,
			)
			stateTopic = fmt.Sprintf(
				"sensorpi/%s/%s/state",
				h.nodeId,
				cParams.ID,
			)
			payload = map[string]any{
				"platform":     "sensor",
				"unique_id":    cParams.ID,
				"name":         cParams.Name,
				"device_class": cParams.Class,
				"state_topic":  stateTopic,
				"device":       h.device,
			}
		)
		if cParams.UnitOfMeasurement != "" {
			payload["unit_of_measurement"] = cParams.UnitOfMeasurement
		}
		if cParams.SuggestedDisplayPrecision != "" {
			payload["suggested_display_precision"] = cParams.SuggestedDisplayPrecision
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		if t := h.client.Publish(topic, 0, true, b); t.Wait() && t.Error() != nil {
			return nil, t.Error()
		}
		return &outputDataSensor{
			topic: stateTopic,
		}, nil
	case typeTrigger:
		cParams := &outputParamsTrigger{}
		if err := params.Parameters.Decode(cParams); err != nil {
			return nil, err
		}
		if cParams.Type == "" {
			cParams.Type = "action"
		}
		var (
			topic = fmt.Sprintf(
				"homeassistant/device_automation/%s/%s_%s/config",
				h.nodeId,
				cParams.Type,
				cParams.Subtype,
			)
			payload = map[string]any{
				"automation_type": "trigger",
				"type":            cParams.Type,
				"subtype":         cParams.Subtype,
				"payload":         cParams.Subtype,
				"topic":           h.actionTopic,
				"device":          h.device,
			}
		)
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		if t := h.client.Publish(topic, 0, true, b); t.Wait() && t.Error() != nil {
			return nil, t.Error()
		}
		return &outputDataTrigger{
			subtype: cParams.Subtype,
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized type \"%s\"", params.Type)
	}
}

func (o *outputDataSensor) Write(h *HomeAssistant, v float64) error {
	if t := h.client.Publish(
		o.topic,
		0,
		true,
		fmt.Sprintf("%f", v),
	); t.Wait() && t.Error() != nil {
		return t.Error()
	}
	return nil
}

func (o *outputDataTrigger) Write(h *HomeAssistant, v float64) error {
	if v == 0 {
		return nil
	}
	if t := h.client.Publish(
		h.actionTopic,
		0,
		false,
		o.subtype,
	); t.Wait() && t.Error() != nil {
		return t.Error()
	}
	return nil
}

func (h *HomeAssistant) Write(data any, v float64) error {
	return data.(outputData).Write(h, v)
}

func (h *HomeAssistant) WriteClose(data any) {}

func (h *HomeAssistant) WatchInit(node *yaml.Node) (any, error) {
	params := &outputTriggerParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	switch params.Type {
	case typeLight:
		cParams := &triggerParamsLight{}
		if err := params.Parameters.Decode(cParams); err != nil {
			return nil, err
		}
		var (
			topic = fmt.Sprintf(
				"homeassistant/light/%s/%s/config",
				h.nodeId,
				cParams.ID,
			)
			commandTopic = fmt.Sprintf(
				"sensorpi/%s/%s/switch",
				h.nodeId,
				cParams.ID,
			)
			payload = map[string]any{
				"platform":      "light",
				"unique_id":     cParams.ID,
				"name":          cParams.Name,
				"command_topic": commandTopic,
				"device":        h.device,
			}
		)
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		if t := h.client.Publish(topic, 0, true, b); t.Wait() && t.Error() != nil {
			return nil, t.Error()
		}
		cmdChan := make(chan bool)
		if t := h.client.Subscribe(
			commandTopic,
			0,
			func(client mqtt.Client, msg mqtt.Message) {
				switch string(msg.Payload()) {
				case "ON":
					cmdChan <- true
				case "OFF":
					cmdChan <- false
				}
			},
		); t.Wait() && t.Error() != nil {
			return nil, t.Error()
		}
		return &triggerDataLight{
			topic:   commandTopic,
			cmdChan: cmdChan,
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized type \"%s\"", params.Type)
	}
}

func (tr *triggerDataLight) Watch(h *HomeAssistant, ctx context.Context) (float64, error) {
	select {
	case v := <-tr.cmdChan:
		if v {
			return 1, nil
		} else {
			return 0, nil
		}
	case <-ctx.Done():
		return 0, context.Canceled
	}
}

func (h *HomeAssistant) Watch(data any, ctx context.Context) (float64, error) {
	return data.(triggerData).Watch(h, ctx)
}

func (tr *triggerDataLight) Close(h *HomeAssistant) {
	if t := h.client.Unsubscribe(tr.topic); t.Wait() && t.Error() != nil {
		log.Warn().Msgf("mqtt: %s", t.Error())
	}
}

func (h *HomeAssistant) WatchClose(data any) {
	data.(triggerData).Close(h)
}

func (h *HomeAssistant) Close() {
	h.client.Disconnect(1000)
}
