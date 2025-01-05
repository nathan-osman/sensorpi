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

type inputData struct {
	W []byte
	R []byte
}

func init() {
	plugin.Register("grove-moisture", func(node *yaml.Node) (any, error) {
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

func (m *Moisture) ReadInit(node *yaml.Node) (any, error) {
	params := &inputParams{}
	if err := node.Decode(params); err != nil {
		return nil, err
	}
	return &inputData{
		W: []byte{byte(0x20 + params.Channel)},
		R: make([]byte, 2),
	}, nil
}

func (m *Moisture) Read(data any) (float64, error) {
	d := data.(*inputData)
	if err := m.conn.Tx(d.W, d.R); err != nil {
		return 0, err
	}
	v := binary.LittleEndian.Uint16(d.R)
	return float64(v), nil
}

func (m *Moisture) ReadClose(any) {}
