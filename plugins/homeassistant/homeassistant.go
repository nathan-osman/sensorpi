package homeassistant

import (
	"encoding/json"
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

const (
	typeSensor  = "sensor"
	typeTrigger = "trigger"
)

// HomeAssistant uses MQTT (with discovery) to interact with Home Assistant
type HomeAssistant struct {
	client      mqtt.Client
	nodeId      string
	actionTopic string
	stateTopic  string
	device      map[string]any
}

type pluginParams struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	NodeId   string `yaml:"node_id"`
}

type outputParams struct {
	Type       string    `yaml:"type"`
	Parameters yaml.Node `yaml:"parameters"`
}

type outputParamsSensor struct {
	ID                string `yaml:"id"`
	Name              string `yaml:"name"`
	UnitOfMeasurement string `yaml:"unit_of_measurement"`
}

type outputParamsTrigger struct {
	Type    string `yaml:"type"`
	Subtype string `yaml:"subtype"`
}

type outputData interface {
	Write(*HomeAssistant, float64) error
}

type outputDataSensor struct {
	//...
}

type outputDataTrigger struct {
	subtype string
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
			stateTopic: fmt.Sprintf(
				"sensorpi/%s/state",
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
	params := &outputParams{}
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
			payload = map[string]any{
				"component":           "sensor",
				"unique_id":           cParams.ID,
				"name":                cParams.Name,
				"unit_of_measurement": cParams.UnitOfMeasurement,
				"state_topic":         h.stateTopic,
				"device":              h.device,
			}
		)
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		if t := h.client.Publish(topic, 0, true, b); t.Wait() && t.Error() != nil {
			return nil, t.Error()
		}
		return &outputDataSensor{}, nil
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

func (h *HomeAssistant) Close() {
	h.client.Disconnect(1000)
}
