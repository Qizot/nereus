package rtmp

import (
	"limen/internal/rtmp/amf"
)

type AnonymousMessage struct {
	Propries []interface{}
	TxId     *float64
	Name     string
}

func (c *AnonymousMessage) Serialize() []byte {
	payload := []interface{}{
		c.Name,
		c.TxId,
	}

	for _, p := range c.Propries {
		payload = append(payload, p)
	}

	bytes, _ := amf.NewAMF0Encoder().Encode(payload)
	return bytes
}

func (c *AnonymousMessage) Deserialize(payload interface{}) error {
	if p, ok := payload.([]interface{}); ok {
		if len(p) < 3 {
			return InvalidMessageFormatErr
		}

		if name, ok := p[0].(string); ok {
			c.Name = name
		} else {
			return InvalidMessageFormatErr
		}

		if txId, ok := p[1].(float64); ok {
			c.TxId = &txId
		} else if p[1] == nil {
			c.TxId = nil
		} else {
			return InvalidMessageFormatErr
		}

		c.Propries = make([]interface{}, 0)

		for _, p := range p[2:] {
			c.Propries = append(c.Propries, p)
		}
	} else {
		return InvalidMessageFormatErr
	}

	return nil
}
