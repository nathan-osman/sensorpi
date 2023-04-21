package moisture

import (
	"encoding/binary"

	"github.com/nathan-osman/sensorpi/plugin"
	"gopkg.in/yaml.v3"
	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// Moisture communicates with the Seeed Studio moisture sensor using the I2C
// bus.
type Moisture struct {
	conn conn.Conn
}

type inputParams struct {
	Channel int `yaml:"channel"`
}

func init() {
	plugin.Register("grove-moisture", func(node *yaml.Node) (plugin.Plugin, error) {
		_, err := host.Init()
		if err != nil {
			return nil, err
		}
		b, err := i2creg.Open("1")
		if err != nil {
			return nil, err
		}
		c := &i2c.Dev{
			Addr: 0x08,
			Bus:  b,
		}
		return &Moisture{
			conn: c,
		}, nil
	})
}

func (m *Moisture) IsInput() bool {
	return true
}

func (m *Moisture) IsOutput() bool {
	return false
}

func (m *Moisture) Read(node *yaml.Node) (float64, error) {
	params := &inputParams{}
	if err := node.Decode(params); err != nil {
		return 0, err
	}
	var (
		w = []byte{0x10 + byte(params.Channel)}
		r = make([]byte, 2)
	)
	if err := m.conn.Tx(w, r); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint16(r)
	return float64(v), nil
}

func (m *Moisture) Write(float64, *yaml.Node) error {
	return nil
}

func (m *Moisture) Close() {}
