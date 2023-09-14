package amf

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

type KeyValuePair struct {
	Key   string
	Value interface{}
}

var (
	NotEnoughDataErr     = errors.New("Not enough data")
	UnknownAMF0MarkerErr = errors.New("Unknown AMF0 marker")
)

type amf0Decoder struct {
}

func NewAMF0Decoder() *amf0Decoder {
	return &amf0Decoder{}
}

func (d *amf0Decoder) Decode(buffer *bufio.Reader) (interface{}, error) {
	marker, err := buffer.ReadByte()
	if err != nil {
		return nil, err
	}

	switch marker {
	case NumberType:
		return d.decodeNumber(buffer)
	case BooleanType:
		return d.decodeBoolean(buffer)
	case StringType:
		return d.decodeString(buffer)
	case KeyValueObjectType:
		return d.decodeKeyValueObject(buffer)
	case NullType:
		return nil, nil
	case ECMAArrayType:
		return d.decodeECMAArray(buffer)
	default:
		return nil, UnknownAMF0MarkerErr
	}
}

func (d *amf0Decoder) decodeNumber(buffer *bufio.Reader) (float64, error) {
	var buff [8]byte
	_, err := io.ReadFull(buffer, buff[:])
	if err != nil {
		return 0.0, NotEnoughDataErr
	}

	return math.Float64frombits(binary.BigEndian.Uint64(buff[:])), nil
}

func (d *amf0Decoder) decodeBoolean(buffer *bufio.Reader) (bool, error) {
	b, err := buffer.ReadByte()
	if err != nil {
		return false, NotEnoughDataErr
	}
	return b == 0x01, nil
}

func (d *amf0Decoder) decodeString(buffer *bufio.Reader) (string, error) {
	var buff [2]byte
	_, err := io.ReadFull(buffer, buff[:])
	if err != nil {
		return "", NotEnoughDataErr
	}

	payload := make([]byte, int(binary.BigEndian.Uint16(buff[:])))
	_, err = io.ReadFull(buffer, payload)
	if err != nil {
		return "", NotEnoughDataErr
	}
	return string(payload), nil
}

func (d *amf0Decoder) decodeKeyValueObject(buffer *bufio.Reader) (map[string]interface{}, error) {
	pairs, err := d.decodeKeyValuePairs(buffer)
	if err != nil {
		return nil, err
	}

	payload := make(map[string]interface{})

	for _, pair := range pairs {
		payload[pair.Key] = pair.Value
	}

	return payload, nil
}

func (d *amf0Decoder) decodeECMAArray(buffer *bufio.Reader) ([]*KeyValuePair, error) {
	var buff [4]byte
	_, err := io.ReadFull(buffer, buff[:])
	if err != nil {
		return nil, NotEnoughDataErr
	}

	return d.decodeKeyValuePairs(buffer)
}

func (d *amf0Decoder) decodeKeyValuePairs(buffer *bufio.Reader) ([]*KeyValuePair, error) {
	payload := make([]*KeyValuePair, 0)

	for {
		endMarker, err := buffer.Peek(3)
		if err != nil {
			return nil, NotEnoughDataErr
		}
		if bytes.Equal(endMarker[:], ObjectEndMarker[:]) {
			return payload, nil
		}

		key, err := d.decodeString(buffer)
		if err != nil {
			return nil, err
		}
		if key == "" {
			break
		}
		value, err := d.Decode(buffer)
		if err != nil {
			return nil, err
		}
		payload = append(payload, &KeyValuePair{Key: key, Value: value})
	}

	return payload, nil
}
