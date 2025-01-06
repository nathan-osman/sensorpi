package influxdb

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
)

// InfluxDB maintains a connection to an InfluxDB server.
type InfluxDB struct {
	client influxdb2.Client
	api    api.WriteAPIBlocking
}

type pluginParams struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type outputParams struct {
	Name string            `yaml:"name"`
	Tags map[string]string `yaml:"tags"`
}

func init() {
	plugin.Register("influxdb", func(node *yaml.Node) (plugin.Plugin, error) {
		params := &pluginParams{}
		if err := node.Decode(params); err != nil {
			return nil, err
		}
		var (
			client = influxdb2.NewClient(
				params.URL,
				fmt.Sprintf("%s:%s", params.Username, params.Password),
			)
			api = client.WriteAPIBlocking("", params.Database)
		)
		i := &InfluxDB{
			client: client,
			api:    api,
		}
		return i, nil
	})
}

func (i *InfluxDB) WriteInit(node *yaml.Node) (any, error) {
	params := &outputParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	return params, nil
}

func (i *InfluxDB) Write(data any, v float64) error {
	var (
		params = data.(*outputParams)
		p      = influxdb2.NewPoint(
			params.Name,
			params.Tags,
			map[string]interface{}{
				"value": v,
			},
			time.Now(),
		)
	)
	return i.api.WritePoint(context.Background(), p)
}

func (i *InfluxDB) WriteClose(any) {}

func (i *InfluxDB) Close() {
	i.client.Close()
}
