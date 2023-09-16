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
	Value interface{}
	Key   string
}

var (
	ErrNotEnoughData     = errors.New("not enough data")
	ErrUnknownAMF0Marker = errors.New("unknown AMF0 marker")
	ErrAmfBufferEmpty    = errors.New("amf buffer empty")
)

type amf0Decoder struct{}

func NewAMF0Decoder() *amf0Decoder {
	return &amf0Decoder{}
}

func (d *amf0Decoder) Decode(buffer *bufio.Reader) ([]interface{}, error) {
	items := make([]interface{}, 0)

	for {
		item, err := d.decodeItem(buffer)

		if errors.Is(err, ErrAmfBufferEmpty) {
			return items, nil
		} else if err != nil {
			return nil, err
		} else {
			items = append(items, item)
		}
	}
}

func (d *amf0Decoder) decodeItem(buffer *bufio.Reader) (interface{}, error) {
	marker, err := buffer.ReadByte()
	if err != nil {
		return nil, ErrAmfBufferEmpty
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
		return nil, ErrUnknownAMF0Marker
	}
}

func (d *amf0Decoder) decodeNumber(buffer *bufio.Reader) (float64, error) {
	var buff [8]byte
	_, err := io.ReadFull(buffer, buff[:])
	if err != nil {
		return 0.0, ErrNotEnoughData
	}

	return math.Float64frombits(binary.BigEndian.Uint64(buff[:])), nil
}

func (d *amf0Decoder) decodeBoolean(buffer *bufio.Reader) (bool, error) {
	b, err := buffer.ReadByte()
	if err != nil {
		return false, ErrNotEnoughData
	}
	return b == 0x01, nil
}

func (d *amf0Decoder) decodeString(buffer *bufio.Reader) (string, error) {
	var buff [2]byte
	_, err := io.ReadFull(buffer, buff[:])
	if err != nil {
		return "", ErrNotEnoughData
	}

	payload := make([]byte, int(binary.BigEndian.Uint16(buff[:])))

	_, err = io.ReadFull(buffer, payload)
	if err != nil {
		return "", ErrNotEnoughData
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
		return nil, ErrNotEnoughData
	}

	return d.decodeKeyValuePairs(buffer)
}

func (d *amf0Decoder) decodeKeyValuePairs(buffer *bufio.Reader) ([]*KeyValuePair, error) {
	payload := make([]*KeyValuePair, 0)

	for {
		endMarker, err := buffer.Peek(3)
		if err != nil {
			return nil, ErrNotEnoughData
		}
		if bytes.Equal(endMarker[:], ObjectEndMarker[:]) {
			_, _ = buffer.Discard(3)
			return payload, nil
		}

		key, err := d.decodeString(buffer)
		if err != nil {
			return nil, err
		}
		if key == "" {
			break
		}
		value, err := d.decodeItem(buffer)
		if err != nil {
			return nil, err
		}
		payload = append(payload, &KeyValuePair{Key: key, Value: value})
	}

	return payload, nil
}
