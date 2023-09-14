package rtmp

import "errors"

var (
	NotEnoughDataErr           = errors.New("Not enough data")
	InvalidHeaderTypeErr       = errors.New("Invalid header type")
	OtherHeaderTypeExpectedErr = errors.New("OtherHeaderTypeExpectedErr")
	InvalidMessageFormatErr    = errors.New("Invalid message format")
)
