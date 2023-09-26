package aac

import (
	"errors"

	"limen/internal/codec"
)

const DefaultSamplesPerFrame = 1024

type ParserOptions struct {
	SamplesPerFrame     uint32
	OutputEncapsulation Encapsulation
}

type parser struct {
	options ParserOptions
}

func NewParser() *parser {
	return &parser{}
}

func (p *parser) Parse(data []byte) (*codec.Frame, error) {
	return nil, errors.New("not Encapsulation")
}
