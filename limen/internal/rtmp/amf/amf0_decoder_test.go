package amf

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)



func TestDecodeAMF0Number(t *testing.T) {
  payload := []byte{
    // type
    0x00,
    // payload
    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
  }

  decoder := NewAMF0Decoder()

  buffer := bufio.NewReader(bytes.NewReader(payload))

  number, err := decoder.Decode(buffer)
  assert.Nil(t, err)
  assert.Equal(t, number, float64(0))
}

func TestDecodeAMF0Boolean(t *testing.T) {
  payload := []byte{
    // type
    0x01,
    // payload
    0x01,
  }
  decoder := NewAMF0Decoder()
  buffer := bufio.NewReader(bytes.NewReader(payload))
  boolean, err := decoder.Decode(buffer)
  assert.Nil(t, err)
  assert.Equal(t, boolean, true)

  payload = []byte{
    // type
    0x01,
    // payload
    0x00,
  }

  buffer = bufio.NewReader(bytes.NewReader(payload))
  boolean, err = decoder.Decode(buffer)
  assert.Nil(t, err)
  assert.Equal(t, boolean, false)
}

func TestDecodeAMF0String(t *testing.T) {
  payload := []byte{
    // type
    0x02,
    // payload
    0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
  }
  decoder := NewAMF0Decoder()
  buffer := bufio.NewReader(bytes.NewReader(payload))
  string, err := decoder.Decode(buffer)
  assert.Nil(t, err)
  assert.Equal(t, string, "hello")

  payload = []byte{
    // type
    0x02,
    // payload
    0x00, 0x00,
  }
  buffer = bufio.NewReader(bytes.NewReader(payload))
  string, err = decoder.Decode(buffer)
  assert.Nil(t, err)
  assert.Equal(t, string, "")
}

func TestDecodeAMF0Object(t *testing.T) {
  // an empty object
  payload := []byte{
    // type
    0x03,
    // payload - end marker
    0x00, 0x00, 0x09,
  }
  decoder := NewAMF0Decoder()
  buffer := bufio.NewReader(bytes.NewReader(payload))
  object, err := decoder.Decode(buffer)
  assert.Nil(t, err)

  if _, ok := object.(map[string]interface{}); !ok {
    t.Errorf("expected object to be of type map[string]interface{}")
  }

  assert.Equal(t, object, map[string]interface{}{})

  // object with simple key-value string pairs
  payload = []byte{
    // type
    0x03,
    // key 1
    0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
    // value 1 type
    0x02,
    // value 1
    0x00, 0x05, 0x77, 0x6f, 0x72, 0x6c, 0x64,
    // end marker
    0x00, 0x00, 0x09,
  }

  buffer = bufio.NewReader(bytes.NewReader(payload))
  object, err = decoder.Decode(buffer)
  assert.Nil(t, err)

  if object, ok := object.(map[string]interface{}); !ok {
    t.Errorf("expected object to be of type map[string]interface{}")
  } else {
    assert.Equal(t, object["hello"], "world")
  }
}

func TestDecodeAMF0Null(t *testing.T) {
  // an empty object
  payload := []byte{
    // type
    0x05,
  }
  decoder := NewAMF0Decoder()
  buffer := bufio.NewReader(bytes.NewReader(payload))
  null, err := decoder.Decode(buffer)
  assert.Nil(t, err)
  assert.Equal(t, null, nil)
}

func TestDecodeAMF0ECMAArray(t *testing.T) {
  // an empty object
  payload := []byte{
    // type
    0x08,
    // size
    0x00, 0x00, 0x00, 0x00,
    // payload - end marker
    0x00, 0x00, 0x09,
  }
  decoder := NewAMF0Decoder()
  buffer := bufio.NewReader(bytes.NewReader(payload))
  object, err := decoder.Decode(buffer)
  assert.Nil(t, err)

  if _, ok := object.([]*KeyValuePair); !ok {
    t.Errorf("expected object to be of type []*KeyValuePair")
  }

  assert.Equal(t, object, []*KeyValuePair{})

  // object with simple key-value string pairs
  payload = []byte{
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
  }

  buffer = bufio.NewReader(bytes.NewReader(payload))
  object, err = decoder.Decode(buffer)
  assert.Nil(t, err)

  if pairs, ok := object.([]*KeyValuePair); !ok {
    t.Errorf("expected object to be of type []*KeyValuePair")
  } else {
    assert.Equal(t, len(pairs), 1)
    assert.Equal(t, pairs[0].Key, "hello")
    assert.Equal(t, pairs[0].Value, "world")
  }
}
