package amf

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeAMF0Number(t *testing.T) {
	encoder := NewAMF0Encoder()

	encoded, err := encoder.Encode(float64(15))

	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, []byte{0x00, 0x40, 0x2e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}))
}

func TestEncodeAMF0Boolean(t *testing.T) {
	encoder := NewAMF0Encoder()

	encoded, err := encoder.Encode(true)

	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, []byte{0x01, 0x01}))

	encoded, err = encoder.Encode(false)

	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, []byte{0x01, 0x00}))
}

func TestEncodeAMF0String(t *testing.T) {
	encoder := NewAMF0Encoder()
	encoded, err := encoder.Encode("test")
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, []byte{0x02, 0x00, 0x04, 0x74, 0x65, 0x73, 0x74}))

	encoded, err = encoder.Encode("")
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, []byte{0x02, 0x00, 0x00}))
}

func TestEncodeAMF0Null(t *testing.T) {
	encoder := NewAMF0Encoder()
	encoded, err := encoder.Encode(nil)
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, []byte{0x05}))
}

func TestEncodeAMF0Object(t *testing.T) {
	encoder := NewAMF0Encoder()
	encoded, err := encoder.Encode(map[string]interface{}{"hello": 0.0})
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, []byte{
		// type
		0x03,
		// key 1
		0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
		// value 1 type
		0x00,
		// value 1
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		// end marker
		0x00, 0x00, 0x09,
	}))

	encoded, err = encoder.Encode(map[string]interface{}{})
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, []byte{
		// type
		0x03,
		// end marker
		0x00, 0x00, 0x09,
	}))
}

func TestEncodeAMF0ECMAArray(t *testing.T) {
	encoder := NewAMF0Encoder()
	encoded, err := encoder.Encode([]*KeyValuePair{
		{"hello", "world"},
	})

	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, []byte{
		// type
		0x08,
		// size
		0x00, 0x00, 0x00, 0x01,
		// key 1
		0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
		// value 1 type
		0x02,
		// value 1
		0x00, 0x05, 0x77, 0x6f, 0x72, 0x6c, 0x64,
		// end marker
		0x00, 0x00, 0x09,
	}))

	encoded, err = encoder.Encode([]*KeyValuePair{})
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, []byte{
		// type
		0x08,
    0x00, 0x00, 0x00, 0x00,
		// end marker
		0x00, 0x00, 0x09,
	}))
}
