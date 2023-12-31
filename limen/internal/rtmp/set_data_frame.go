package rtmp

import (
	"github.com/mitchellh/mapstructure"

	"limen/internal/rtmp/amf"
)

type SetDataFrameMessage struct {
	Encoder         string
	Duration        float64
	FileSize        float64 `mapstructure:"filesize"`
	Width           float64
	Height          float64
	VideoCodecId    float64 `mapstructure:"videocodecid"`
	VideoDataRate   float64 `mapstructure:"videodatarate"`
	Framerate       float64
	AudioCodecId    float64 `mapstructure:"audiocodecid"`
	AudioSampleRate float64 `mapstructure:"audiosamplerate"`
	AudioSampleSize float64 `mapstructure:"audiosamplesize"`
	Stereo          bool
}

func (c *SetDataFrameMessage) Serialize() []byte {
	payload := map[string]interface{}{
		"encoder":         c.Encoder,
		"duration":        c.Duration,
		"filesize":        c.FileSize,
		"width":           c.Width,
		"height":          c.Height,
		"videocodecid":    c.VideoCodecId,
		"videoDataRate":   c.VideoDataRate,
		"framerate":       c.Framerate,
		"audioCodecId":    c.AudioCodecId,
		"audioSampleRate": c.AudioSampleRate,
		"audioSampleSize": c.AudioSampleSize,
		"stereo":          c.Stereo,
	}

	bytes, _ := amf.NewAMF0Encoder().Encode([]interface{}{payload})

	return bytes
}

func (c *SetDataFrameMessage) Deserialize(payload interface{}) error {
	if p, ok := payload.([]interface{}); ok {
		if len(p) != 3 {
			return ErrInvalidMessageFormat
		}

		if p[0] != "@setDataFrame" {
			return ErrInvalidMessageFormat
		}

		if p[1] != "onMetaData" {
			return ErrInvalidMessageFormat
		}

		fields := map[string]interface{}{}

		for _, pair := range p[2].([]*amf.KeyValuePair) {
			fields[pair.Key] = pair.Value
		}

		if err := mapstructure.Decode(fields, c); err != nil {
			return ErrInvalidMessageFormat
		}
	} else {
		return ErrInvalidMessageFormat
	}
	return nil
}
