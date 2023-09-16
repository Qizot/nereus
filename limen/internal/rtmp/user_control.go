package rtmp

type UserControlMessage struct {
	Data      []byte
	EventType uint16
}

func (c *UserControlMessage) Type() uint8 {
	return UserControlType
}

func (c *UserControlMessage) Serialize() []byte {
	payload := make([]byte, 2+len(c.Data))
	payload[0] = byte(c.EventType >> 8)
	payload[1] = byte(c.EventType & 0xFF)

	copy(payload[2:], c.Data)

	return payload
}

func (c *UserControlMessage) Deserialize(data []byte) error {
	if len(data) < 2 {
		return ErrInvalidMessageFormat
	}

	c.EventType = uint16(data[0])<<8 | uint16(data[1])
	c.Data = make([]byte, len(data[2:]))
	copy(c.Data, data[2:])

	return nil
}
