package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"net"

	"limen/internal/flv"
	"limen/internal/rtmp"
)

func main() {
	logger := slog.Default()
	rtmpServer := &rtmp.RtmpServer{Host: "0.0.0.0", Port: 1935, Logger: logger, Handler: func(conn net.Conn) error {
		callbacks := &rtmp.HandlerCallabcks{
			OnAuthorize: func(streamKey string) bool {
				return true
			},
			OnSetDataFrame: func(message rtmp.SetDataFrameMessage) bool {
				return true
			},
		}

		mediaStream := make(chan interface{})
		handler := rtmp.NewHandler(conn, logger, callbacks, mediaStream)

		go runFlvReader(logger, mediaStream)

		err := handler.Run()
		if err != nil {
			fmt.Printf("Error running handler %+v\n", err)
		}
		return err
	}}

	rtmpServer.Run()
}

func runFlvReader(logger *slog.Logger, mediaStream chan interface{}) {
	decoder := flv.NewFlvDecoder()

	buffer := &bytes.Buffer{}
	writer := bufio.NewWriter(buffer)

	reader := bufio.NewReader(buffer)

	for msg := range mediaStream {
		switch message := msg.(type) {
		case rtmp.MediaStreamInfo:
			fmt.Printf("Media info %+v\n", message)

		case rtmp.MediaStreamData:
			writer.Write(message.Data)
			writer.Flush()

			packet, err := decoder.Decode(reader)
			if err != nil {
				logger.Error(fmt.Sprintf("Error -> %+v\n", err))
			}

			if packet != nil {
				packet.Data = []byte{}
				if packet.Type == 1 {
					logger.Info(fmt.Sprintf("Got FLV packet %+v\n", packet))
				}
			}

		}
	}
}
