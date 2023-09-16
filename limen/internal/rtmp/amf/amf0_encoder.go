package amf

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

var UnsupportedAMFTypeErr = errors.New("unsupported AMF type")

type amf0Encoder struct{}

func NewAMF0Encoder() *amf0Encoder {
	return &amf0Encoder{}
}

func (e *amf0Encoder) Encode(value interface{}) ([]byte, error) {
	switch value := value.(type) {
	case float64:
		return encodeFloat(value), nil
	case bool:
		return encodeBool(value), nil
	case string:
		return encodeString(value), nil
	case nil:
		return encodeNil(), nil
	case map[string]interface{}:
		return e.encodeMap(value)
	case []*KeyValuePair:
		return e.encodeArray(value)
	}

	return nil, UnsupportedAMFTypeErr
}

func encodeFloat(value float64) []byte {
	buff := make([]byte, 9)
	buff[0] = NumberType
	binary.BigEndian.PutUint64(buff[1:], math.Float64bits(value))

	return buff
}

func encodeBool(value bool) []byte {
	buff := make([]byte, 2)
	buff[0] = BooleanType
	if value {
		buff[1] = 0x01
	} else {
		buff[1] = 0x00
	}

	return buff
}

func encodeString(value string) []byte {
	buff := make([]byte, 3)
	buff[0] = StringType
	binary.BigEndian.PutUint16(buff[1:], uint16(len(value)))

	return append(buff, []byte(value)...)
}

func encodeRawString(value string) []byte {
	buff := make([]byte, 2)
	binary.BigEndian.PutUint16(buff[:], uint16(len(value)))

	return append(buff, []byte(value)...)
}

func encodeNil() []byte {
	return []byte{NullType}
}

func (e *amf0Encoder) encodeMap(valueMap map[string]interface{}) ([]byte, error) {
	buff := new(bytes.Buffer)
	writer := bufio.NewWriter(buff)
	writer.WriteByte(KeyValueObjectType)

	for key, value := range valueMap {
		_, err := writer.Write(encodeRawString(key))
		if err != nil {
			return nil, err
		}

		valueBytes, err := e.Encode(value)
		if err != nil {
			return nil, err
		}

		_, err = writer.Write(valueBytes)
		if err != nil {
			return nil, err
		}
	}

	_, err := writer.Write(ObjectEndMarker[:])
	if err != nil {
		return nil, err
	}

	err = writer.Flush()
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (e *amf0Encoder) encodeArray(value []*KeyValuePair) ([]byte, error) {
	buff := new(bytes.Buffer)
	writer := bufio.NewWriter(buff)
	writer.WriteByte(ECMAArrayType)

	var sizeBuff [4]byte
	binary.BigEndian.PutUint32(sizeBuff[:], uint32(len(value)))

	_, err := writer.Write(sizeBuff[:])
	if err != nil {
		return nil, err
	}

	for _, pair := range value {
		_, err := writer.Write(encodeRawString(pair.Key))
		if err != nil {
			return nil, err
		}

		valueBytes, err := e.Encode(pair.Value)
		if err != nil {
			return nil, err
		}

		_, err = writer.Write(valueBytes)
		if err != nil {
			return nil, err
		}
	}

	_, err = writer.Write(ObjectEndMarker[:])
	if err != nil {
		return nil, err
	}

	err = writer.Flush()
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
