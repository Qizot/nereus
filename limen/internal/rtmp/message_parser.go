package rtmp

import (
	"bufio"
	"bytes"

	"limen/internal/rtmp/amf"
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

		return msg, nil

	case AmfDataType, AmfCommandType:
		return parseAmfMessage(message.Payload)

	default:
		return nil, ErrInvalidHeaderType
	}

	return nil, ErrInvalidHeaderType
}

func parseAmfMessage(payload []byte) (interface{}, error) {
	buffer := bufio.NewReader(bytes.NewReader(payload))
	if data, err := amf.NewAMF0Decoder().Decode(buffer); err != nil {
		return nil, err
	} else {
		if len(data) < 1 {
			return nil, ErrInvalidMessageFormat
		}

		if name, ok := data[0].(string); ok {
			switch name {
			case "connect":
				msg := &ConnectCommand{}

				if err := msg.Deserialize(data); err != nil {
					return nil, err
				}
				return msg, nil

			case "releaseStream":
				msg := &ReleaseStreamCommand{}

				if err := msg.Deserialize(data); err != nil {
					return nil, err
				}
				return msg, nil

			case "FCPublish":
				msg := &FCPublishCommand{}

				if err := msg.Deserialize(data); err != nil {
					return nil, err
				}
				return msg, nil

			case "createStream":
				msg := &CreateStreamCommand{}

				if err := msg.Deserialize(data); err != nil {
					return nil, err
				}
				return msg, nil

			case "publish":
				msg := &PublishCommand{}

				if err := msg.Deserialize(data); err != nil {
					return nil, err
				}
				return msg, nil

			case "@setDataFrame":
				msg := &SetDataFrameMessage{}

				if err := msg.Deserialize(data); err != nil {
					return nil, err
				}
				return msg, nil
			default:
				return nil, ErrInvalidMessageFormat
			}
		} else {
			return nil, ErrInvalidMessageFormat
		}
	}
}
