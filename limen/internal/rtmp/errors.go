package rtmp

import "errors"

var (
	ErrNotEnoughData           = errors.New("not enough data")
	ErrInvalidHeaderType       = errors.New("invalid header type")
	ErrOtherHeaderTypeExpected = errors.New("ErrOtherHeaderTypeExpected")
	ErrInvalidMessageFormat    = errors.New("invalid message format")
	ErrInvalidHandshake        = errors.New("invalid handshake")
)
