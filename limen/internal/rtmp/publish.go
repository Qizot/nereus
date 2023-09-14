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
			return InvalidMessageFormatErr
		}

		if p[0] != "publish" {
			return InvalidMessageFormatErr
		}

		if txId, ok := p[1].(float64); ok {
			c.TxId = txId
		} else {
			return InvalidMessageFormatErr
		}

		if streamKey, ok := p[3].(string); ok {
			c.StreamKey = streamKey
		} else {
			return InvalidMessageFormatErr
		}

		if publishType, ok := p[4].(string); ok {
			c.PublishType = publishType
		} else {
			return InvalidMessageFormatErr
		}

	} else {
		return InvalidMessageFormatErr
	}

	return nil
}
