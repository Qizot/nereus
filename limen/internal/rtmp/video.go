package rtmp

type VideoMessage struct {
	Data []byte
}

func (c *VideoMessage) Serialize() []byte {
	return c.Data
}

func (c *VideoMessage) Deserialize(data []byte) error {
	c.Data = data
	return nil
}
