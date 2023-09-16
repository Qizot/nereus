package rtmp

import (
	"fmt"
	"log/slog"
	"net"
)

type ConnHandleFunc func(conn net.Conn) error

type RtmpServer struct {
	Handler ConnHandleFunc
	Logger  *slog.Logger
	Host    string
	Port    int
}

func (s *RtmpServer) Run() error {
	if s.Logger == nil {
		s.Logger = slog.Default()
	}

	s.Logger.Info("Starting RTMP server", "host", s.Host, "port", s.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		if conn, err := listener.Accept(); err != nil {
			return err
		} else {
			go s.Handler(conn)
		}
	}
}
