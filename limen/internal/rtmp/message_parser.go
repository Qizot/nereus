package rtmp

import (
	"bufio"
	"bytes"

	"limen/internal/rtmp/amf"
)

const (
	SetChunkSizeType     = 0x1
	UserControlType      = 0x4
	WindowAckSizeType    = 0x5
	SetPeerBandwidthType = 0x6
	AudioType            = 0x8
	VideoType            = 0x9
	AmfDataType          = 0x12
	AmfCommandType       = 0x14
)

func ParseMessage(message *Message) (interface{}, error) {
	switch message.Header.Type {
	case SetChunkSizeType:
		msg := &SetChunkSizeMessage{}

		if err := msg.Deserialize(message.Payload); err != nil {
			return nil, err
		}

		return msg, nil

	case UserControlType:
		msg := &UserControlMessage{}

		if err := msg.Deserialize(message.Payload); err != nil {
			return nil, err
		}

		return msg, nil

	case WindowAckSizeType:
		msg := &WindowAcknowledgementSizeMessage{}

		if err := msg.Deserialize(message.Payload); err != nil {
			return nil, err
		}

		return msg, nil

	case SetPeerBandwidthType:
		msg := &SetPeerBandwidthMessage{}

		if err := msg.Deserialize(message.Payload); err != nil {
			return nil, err
		}

		return msg, nil

	case AudioType:
		msg := &AudioMessage{}

		if err := msg.Deserialize(message.Payload); err != nil {
			return nil, err
		}

		return msg, nil

	case VideoType:
		msg := &VideoMessage{}

		if err := msg.Deserialize(message.Payload); err != nil {
			return nil, err
		}

	case AmfDataType, AmfCommandType:
		return parseAmfMessage(message.Payload)

	default:
		return nil, InvalidHeaderTypeErr
	}

	return nil, InvalidHeaderTypeErr
}

func parseAmfMessage(payload []byte) (interface{}, error) {
	buffer := bufio.NewReader(bytes.NewReader(payload))
	if data, err := amf.NewAMF0Decoder().Decode(buffer); err != nil {
		return nil, err
	} else {
		if params, ok := data.([]interface{}); ok {
			if len(params) < 1 {
				return nil, InvalidMessageFormatErr
			}

			if name, ok := params[0].(string); ok {
				switch name {
				case "connect":
					msg := &ConnectCommand{}

					if err := msg.Deserialize(params[1:]); err != nil {
						return nil, err
					}
					return msg, nil

				case "releaseStream":
					msg := &ReleaseStreamCommand{}

					if err := msg.Deserialize(params[1:]); err != nil {
						return nil, err
					}
					return msg, nil

				case "FCPublish":
					msg := &FCPublishCommand{}

					if err := msg.Deserialize(params[1:]); err != nil {
						return nil, err
					}
					return msg, nil

				case "createStream":
					msg := &CreateStreamCommand{}

					if err := msg.Deserialize(params[1:]); err != nil {
						return nil, err
					}
					return msg, nil

				case "publish":
					msg := &PublishCommand{}

					if err := msg.Deserialize(params[1:]); err != nil {
						return nil, err
					}
					return msg, nil

				case "@setDataFrame":
					msg := &SetDataFrameMessage{}

					if err := msg.Deserialize(params[1:]); err != nil {
						return nil, err
					}
					return msg, nil
				default:
					return nil, InvalidMessageFormatErr
				}
			} else {
				return nil, InvalidMessageFormatErr
			}
		} else {
			return nil, InvalidMessageFormatErr
		}
	}
}
