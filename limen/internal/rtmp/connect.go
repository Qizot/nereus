package rtmp

import (
	"github.com/mitchellh/mapstructure"

	"limen/internal/rtmp/amf"
)

type ConnectCommand struct {
	App            string  `mapstucture:"app"`
	Type           string  `mapstucture:"type"`
	FlashVer       string  `mapstructure:"flashVer"`
	TcUrl          string  `mapstructure:"tcUrl"`
	SupportsGoAway bool    `mapstucture:",omitempty"`
	TxId           float64 `mapstructure:"-"`
}

func (c *ConnectCommand) Serialize() []byte {
	payload := []interface{}{
		"connect",
		c.TxId,
		map[string]interface{}{
			"app":            c.App,
			"type":           c.Type,
			"flashVer":       c.FlashVer,
			"tcUrl":          c.TcUrl,
			"supportsGoAway": c.SupportsGoAway,
		},
	}
	bytes, _ := amf.NewAMF0Encoder().Encode(payload)

	return bytes
}

func (c *ConnectCommand) Deserialize(payload interface{}) error {
	if p, ok := payload.([]interface{}); ok {
		if len(p) != 3 {
			return ErrInvalidMessageFormat
		}

		if p[0] != "connect" {
			return ErrInvalidMessageFormat
		}

		if err := mapstructure.Decode(p[2], c); err != nil {
			return ErrInvalidMessageFormat
		}

		if txId, ok := p[1].(float64); ok {
			c.TxId = txId
		} else {
			return ErrInvalidMessageFormat
		}

	} else {
		return ErrInvalidMessageFormat
	}

	return nil
}
