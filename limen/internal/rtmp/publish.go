package rtmp

import (
	"limen/internal/rtmp/amf"
)

type PublishCommand struct {
	StreamKey   string
	PublishType string
	TxId        float64
}

func (c *PublishCommand) Serialize() []byte {
	payload := []interface{}{
		"publish",
		c.TxId,
		nil,
		c.StreamKey,
		c.PublishType,
	}

	bytes, _ := amf.NewAMF0Encoder().Encode(payload)
	return bytes
}

func (c *PublishCommand) Deserialize(payload interface{}) error {
	if p, ok := payload.([]interface{}); ok {
		if len(p) != 5 {
			return ErrInvalidMessageFormat
		}

		if p[0] != "publish" {
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

		if publishType, ok := p[4].(string); ok {
			c.PublishType = publishType
		} else {
			return ErrInvalidMessageFormat
		}

	} else {
		return ErrInvalidMessageFormat
	}

	return nil
}
