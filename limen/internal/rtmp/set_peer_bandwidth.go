package rtmp

import (
	"encoding/binary"
)

type SetPeerBandwidthMessage struct {
	Size uint32
}

func (c *SetPeerBandwidthMessage) Type() uint8 {
	return SetPeerBandwidthType
}

func (c *SetPeerBandwidthMessage) Serialize() []byte {
	payload := make([]byte, 5)
	binary.BigEndian.PutUint32(payload, c.Size)
	payload[4] = 0x02

	return payload
}

func (c *SetPeerBandwidthMessage) Deserialize(data []byte) error {
	if len(data) != 5 {
		return ErrInvalidMessageFormat
	}
	if data[5] != 0x02 {
		return ErrInvalidMessageFormat
	}

	c.Size = binary.LittleEndian.Uint32(data[:4])

	return nil
}
