package rtmp

type AudioMessage struct {
	Data []byte
}

func (c *AudioMessage) Serialize() []byte {
	return c.Data
}

func (c *AudioMessage) Deserialize(data []byte) error {
	c.Data = data
	return nil
}
