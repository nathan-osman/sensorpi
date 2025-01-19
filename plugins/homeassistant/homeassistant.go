package homeassistant

import (
	"encoding/json"
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
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

type outputParams struct {
	Type    string `yaml:"type"`
	Subtype string `yaml:"subtype"`
}

type outputData struct {
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
	if params.Type == "" {
		params.Type = "action"
	}
	var (
		topic = fmt.Sprintf(
			"homeassistant/device_automation/%s/%s_%s/config",
			h.nodeId,
			params.Type,
			params.Subtype,
		)
		payload = map[string]any{
			"automation_type": "trigger",
			"type":            params.Type,
			"subtype":         params.Subtype,
			"payload":         params.Subtype,
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
	return &outputData{
		subtype: params.Subtype,
	}, nil
}

func (h *HomeAssistant) Write(data any, v float64) error {
	params := data.(*outputData)
	if v == 0 {
		return nil
	}
	if t := h.client.Publish(
		h.actionTopic,
		0,
		true,
		params.subtype,
	); t.Wait() && t.Error() != nil {
		return t.Error()
	}
	return nil
}

func (h *HomeAssistant) WriteClose(data any) {}

func (h *HomeAssistant) Close() {
	h.client.Disconnect(1000)
}
