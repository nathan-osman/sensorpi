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

type outputConfig struct {
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
		cfg := &outputConfig{}
		if err := node.Decode(cfg); err != nil {
			return nil, err
		}
		var (
			client = influxdb2.NewClient(
				cfg.URL,
				fmt.Sprintf("%s:%s", cfg.Username, cfg.Password),
			)
			api = client.WriteAPIBlocking("", cfg.Database)
		)
		i := &InfluxDB{
			client: client,
			api:    api,
		}
		return i, nil
	})
}

func (i *InfluxDB) IsInput() bool {
	return false
}

func (i *InfluxDB) IsOutput() bool {
	return true
}

func (i *InfluxDB) Read(*yaml.Node) (float64, error) {
	return 0, nil
}

func (i *InfluxDB) Write(v float64, node *yaml.Node) error {
	params := &outputParams{}
	if err := node.Decode(params); err != nil {
		return err
	}
	p := influxdb2.NewPoint(
		params.Name,
		params.Tags,
		map[string]interface{}{
			"value": v,
		},
		time.Now(),
	)
	return i.api.WritePoint(context.Background(), p)
}

func (i *InfluxDB) Close() {
	i.client.Close()
}
