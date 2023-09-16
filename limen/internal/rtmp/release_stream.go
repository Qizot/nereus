package rtmp

import (
	"limen/internal/rtmp/amf"
)

type ReleaseStreamCommand struct {
	StreamKey string
	TxId      float64
}

func (c *ReleaseStreamCommand) Serialize() []byte {
	payload := []interface{}{
		"releaseStream",
		c.TxId,
		nil,
		c.StreamKey,
	}

	bytes, _ := amf.NewAMF0Encoder().Encode(payload)
	return bytes
}

func (c *ReleaseStreamCommand) Deserialize(payload interface{}) error {
	if p, ok := payload.([]interface{}); ok {
		if len(p) != 4 {
			return ErrInvalidMessageFormat
		}

		if p[0] != "releaseStream" {
			return ErrInvalidMessageFormat
		}

		if txId, ok := p[1].(float64); ok {
			c.TxId = txId
		} else {
			return ErrInvalidMessageFormat
		}

		if streamKey, ok := p[3].(string); ok {
			c.StreamKey = streamKey
		} else {
			return ErrInvalidMessageFormat
		}
	} else {
		return ErrInvalidMessageFormat
	}

	return nil
}
