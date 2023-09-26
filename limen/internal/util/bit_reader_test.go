package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitReaderReturniningProperBits(t *testing.T) {
	data := []byte{
		0b11010000,
		0b10101010,
		0b00110011,
	}

	reader := BitReader{
		Data: data,
	}

	payload := uint64(0)
	// reading first byte
	assert.True(t, reader.ReadBits(2, &payload))
	assert.Equal(t, uint8(0b11), uint8(payload))

	assert.True(t, reader.ReadBits(4, &payload))
	assert.Equal(t, uint8(0b0100), uint8(payload))

	assert.True(t, reader.ReadBits(1, &payload))
	assert.Equal(t, uint8(0b0), uint8(payload))

	assert.True(t, reader.ReadBits(1, &payload))
	assert.Equal(t, uint8(0b0), uint8(payload))

	// reading seconds and thrid byte
	assert.True(t, reader.ReadBits(15, &payload))
	assert.Equal(t, uint16(0b101010100011001), uint16(payload))

	assert.True(t, reader.ReadBits(1, &payload))
	assert.Equal(t, uint8(0b1), uint8(payload))

	// we already read all bits
	assert.False(t, reader.ReadBits(1, &payload))
	assert.Equal(t, 0, int(payload))

	data = []byte{
		0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00,
	}

	reader = BitReader{
		Data: data,
	}

	// read 64 bits
	assert.True(t, reader.ReadBits(64, &payload))
	assert.Equal(t, uint64(0xff00ff00ff00ff00), payload)

	// buffer is empty
	assert.False(t, reader.ReadBits(1, &payload))
	assert.Equal(t, 0, int(payload))

	// reading with some bits available but not enough for the requested amount
	data = []byte{
		0xff,
	}

	reader = BitReader{
		Data: data,
	}

	assert.False(t, reader.ReadBits(9, &payload))
	assert.Equal(t, int(payload), 0)
}

func TestBitReaderSkippingBits(t *testing.T) {
	data := []byte{
		0b11010001,
	}

	reader := BitReader{
		Data: data,
	}

	payload := uint64(0)
	// reading first byte
	assert.True(t, reader.SkipBits(2))
	assert.True(t, reader.ReadBits(2, &payload))
	assert.Equal(t, uint8(0b01), uint8(payload))

	assert.True(t, reader.SkipBits(3))
	assert.True(t, reader.ReadBits(1, &payload))
	assert.Equal(t, uint8(0b1), uint8(payload))

	assert.False(t, reader.SkipBits(3))

	// skipping while there are some bits available but not enough
	data = []byte{
		0xff,
	}

	reader = BitReader{
		Data: data,
	}

	assert.False(t, reader.SkipBits(9))
}

func TestBitReaderBitsAvailable(t *testing.T) {
	data := []byte{
		0b11010001,
	}

	reader := BitReader{
		Data: data,
	}

	var payload uint64
	assert.Equal(t, 8, reader.BitsAvailable())
	assert.True(t, reader.ReadBits(3, &payload))
	assert.Equal(t, 5, reader.BitsAvailable())
	assert.True(t, reader.ReadBits(4, &payload))
	assert.Equal(t, 1, reader.BitsAvailable())
	assert.True(t, reader.ReadBits(1, &payload))
	assert.Equal(t, 0, reader.BitsAvailable())
}

func TestBitReaderReadingSlice(t *testing.T) {
	data := []byte{
		0xff, 0x00, 0xff, 0x00, 0xff, 0xff,
	}

	reader := BitReader{
		Data: data,
	}

	assert.True(t, reader.SkipBits(8))

	var payload [2]byte

	assert.Equal(t, 5*8, reader.BitsAvailable())
	assert.True(t, reader.ReadSlice(payload[:]))
	assert.Equal(t, 3*8, reader.BitsAvailable())

	var bytePayload uint64
	assert.True(t, reader.ReadBits(8, &bytePayload))
	assert.Equal(t, uint8(0x00), uint8(bytePayload))

	var largePayload [10]byte
	assert.False(t, reader.ReadSlice(largePayload[:]))

	reader = BitReader{
		Data: data,
	}

	// don't allow for reading when non 8-bit aligned
	assert.True(t, reader.ReadBits(1, &bytePayload))
	assert.False(t, reader.ReadSlice(payload[:]))
}
