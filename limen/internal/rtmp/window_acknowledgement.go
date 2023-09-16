package rtmp

import (
	"encoding/binary"
)

type WindowAcknowledgementSizeMessage struct {
	Size uint32
}

func (c *WindowAcknowledgementSizeMessage) Type() uint8 {
	return WindowAckSizeType
}

func (c *WindowAcknowledgementSizeMessage) Serialize() []byte {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, c.Size)
	return payload
}

func (c *WindowAcknowledgementSizeMessage) Deserialize(data []byte) error {
	if len(data) != 4 {
		return ErrInvalidMessageFormat
	}
	c.Size = binary.BigEndian.Uint32(data)

	return nil
}
