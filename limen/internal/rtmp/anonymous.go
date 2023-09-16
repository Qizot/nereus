package rtmp

import (
	"bufio"
	"bytes"

	"limen/internal/rtmp/amf"
)

type AnonymousMessage struct {
	Properties []interface{}
	TxId       *float64
	Name       string
}

func (c *AnonymousMessage) Type() uint8 {
	return AmfCommandType
}

func (c *AnonymousMessage) Serialize() []byte {
	payload := []interface{}{
		c.Name,
	}

	if c.TxId != nil {
		payload = append(payload, *(c.TxId))
	}

	for _, p := range c.Properties {
		payload = append(payload, p)
	}

	buff := new(bytes.Buffer)
	writer := bufio.NewWriter(buff)

	for _, item := range payload {
		bytes, err := amf.NewAMF0Encoder().Encode(item)
		if err != nil {
			panic("failed to encode AMF payload payload")
		}

		if _, err := writer.Write(bytes); err != nil {
			panic("failed to writer to amf buffer")
		}
	}

	if err := writer.Flush(); err != nil {
		panic("failed to flush amf buffer")
	}

	return buff.Bytes()
}

func (c *AnonymousMessage) Deserialize(payload interface{}) error {
	if p, ok := payload.([]interface{}); ok {
		if len(p) < 3 {
			return ErrInvalidMessageFormat
		}

		if name, ok := p[0].(string); ok {
			c.Name = name
		} else {
			return ErrInvalidMessageFormat
		}

		if txId, ok := p[1].(float64); ok {
			c.TxId = &txId
		} else if p[1] == nil {
			c.TxId = nil
		} else {
			return ErrInvalidMessageFormat
		}

		c.Properties = make([]interface{}, 0)

		for _, p := range p[2:] {
			c.Properties = append(c.Properties, p)
		}
	} else {
		return ErrInvalidMessageFormat
	}

	return nil
}
