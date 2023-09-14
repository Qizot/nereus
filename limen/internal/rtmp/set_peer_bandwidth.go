package rtmp

import (
	"encoding/binary"
)

type SetPeerBandwidthMessage struct {
	Size uint32
}

func (c *SetPeerBandwidthMessage) Serialize() []byte {
	payload := make([]byte, 5)
	binary.LittleEndian.PutUint32(payload, c.Size)
	payload[4] = 0x02

	return payload
}

func (c *SetPeerBandwidthMessage) Deserialize(data []byte) error {
	if len(data) != 5 {
		return InvalidMessageFormatErr
	}
	if data[5] != 0x02 {
		return InvalidMessageFormatErr
	}

	c.Size = binary.LittleEndian.Uint32(data[:4])

	return nil
}
