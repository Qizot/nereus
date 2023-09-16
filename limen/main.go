package main

import (
	"fmt"
	"net"

	"limen/internal/rtmp"
)

func main() {
	rtmpServer := &rtmp.RtmpServer{Host: "0.0.0.0", Port: 1935, Handler: func(conn net.Conn) error {
		callbacks := &rtmp.HandlerCallabcks{
			OnAuthorize: func(streamKey string) bool {
				return true
			},
		}

		handler := rtmp.NewHandler(conn, callbacks)

		err := handler.Run()
		if err != nil {
			fmt.Printf("Error running handler %+v\n", err)
		}
		return err
	}}

	rtmpServer.Run()
}
