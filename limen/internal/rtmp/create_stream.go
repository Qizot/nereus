package rtmp

import (
	"limen/internal/rtmp/amf"
)

type CreateStreamCommand struct {
	TxId float64
}

func (c *CreateStreamCommand) Serialize() []byte {
	payload := []interface{}{
		"createStream",
		c.TxId,
		nil,
	}

	bytes, _ := amf.NewAMF0Encoder().Encode(payload)
	return bytes
}

func (c *CreateStreamCommand) Deserialize(payload interface{}) error {
	if p, ok := payload.([]interface{}); ok {
		if len(p) != 3 {
			return ErrInvalidMessageFormat
		}

		if p[0] != "createStream" {
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
