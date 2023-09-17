package rtmp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"
)

const (
	WindowAcknowledgementSize = 2_500_000
	PeerBandwidthSize         = 2_5_000_000
)

type HandlerCallabcks struct {
	OnAuthorize    func(streamKey string) bool
	OnSetDataFrame func(message SetDataFrameMessage) bool
}

type handler struct {
	logger              *slog.Logger
	handshakeFinished   bool
	connInitialized     bool
	connAuthorized      bool
	setDataFrameHandled bool
	conn                net.Conn
	callbacks           *HandlerCallabcks
	reader              *bufio.Reader
	writer              *bufio.Writer
	messageReader       *messageReader
	messageWriter       *messageWriter
	currentTxId         uint32
	mediaStreamWrapper  *mediaFlvWrapper
	mediaChannel        chan interface{}
}

func NewHandler(conn net.Conn, logger *slog.Logger, callbacks *HandlerCallabcks, mediaChannel chan interface{}) *handler {
	return &handler{
		conn:               conn,
		logger:             logger,
		callbacks:          callbacks,
		reader:             bufio.NewReader(conn),
		writer:             bufio.NewWriter(conn),
		messageReader:      NewMessageReader(),
		messageWriter:      NewMessageWriter(),
		currentTxId:        1,
		mediaStreamWrapper: NewMediaFlvWrapper(true, true),
		mediaChannel:       mediaChannel,
	}
}

func (h *handler) Run() error {
	defer h.conn.Close()

	defer func() {
		h.logger.Info("Closing connection")
	}()

	h.logger.Info("Handling connection")
	for {
		h.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		if !h.handshakeFinished {
			err := h.handleHandshake()
			if err == ErrNotEnoughData {
				continue
			}

			if err != nil {
				return err
			}

			h.handshakeFinished = true
			continue
		}

		if !h.connInitialized {
			err := h.handleInitialization()

			if err == ErrNotEnoughData {
				continue
			}

			if err != nil {
				h.logger.Info("Failed initialization")
				return err
			}

			h.connInitialized = true
			continue
		}

		if !h.connAuthorized {
			err := h.handleAuthorization()

			if errors.Is(err, ErrNotEnoughData) {
				continue
			}

			if err != nil {
				h.logger.Info("Failed authorization")
				return err
			}

			h.connAuthorized = true
			continue
		}

		if !h.setDataFrameHandled {
			err := h.handleSetDataFrame()
			if err != nil {
				h.logger.Info("Failed to handle setDataFrame")
				return err
			}

			h.setDataFrameHandled = true
			continue
		}

		err := h.handleMessage()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			h.logger.Info(fmt.Sprintf("Failed to handle message: %s", err.Error()))
			return err
		}
	}
}

func (h *handler) handleHandshake() error {
	handshake := NewHandshake()

	if err := handshake.ReceiveC0C1(h.reader); err != nil {
		return err
	}

	s0s1 := handshake.GenerateS0S1()

	if _, err := h.writer.Write(s0s1); err != nil {
		return err
	}

	if err := h.writer.Flush(); err != nil {
		return err
	}

	if err := handshake.ReceiveC2(h.reader); err != nil {
		return err
	}

	s2 := handshake.GetS2()

	if _, err := h.writer.Write(s2); err != nil {
		return err
	}

	if err := h.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (h *handler) handleInitialization() error {
	setChunkSize, err := h.readAndParseMessage()
	if err != nil {
		return err
	}

	if setChunkSize, ok := setChunkSize.(*SetChunkSizeMessage); ok {
		h.messageReader.SetChunkSize(int32(setChunkSize.ChunkSize))
		h.messageWriter.SetChunkSize(int32(setChunkSize.ChunkSize))
	} else {
		return errors.New("expected SetChunkSize message")
	}

	connect, err := h.readAndParseMessage()
	if _, ok := connect.(*ConnectCommand); !ok {
		return errors.New("expected Connect command")
	}

	winAckMsg := &WindowAcknowledgementSizeMessage{
		Size: WindowAcknowledgementSize,
	}

	const defaultChunkStreamId = 2
	if err := h.serializeAndSendMessage(defaultChunkStreamId, winAckMsg); err != nil {
		return err
	}

	setPeerBandMsg := &SetPeerBandwidthMessage{
		Size: PeerBandwidthSize,
	}

	if err := h.serializeAndSendMessage(defaultChunkStreamId, setPeerBandMsg); err != nil {
		return err
	}

	userControlMsg := &UserControlMessage{
		EventType: 0,
		Data:      []byte{0x0, 0x0, 0x0, 0x0},
	}

	if err := h.serializeAndSendMessage(defaultChunkStreamId, userControlMsg); err != nil {
		return err
	}

	setChunkSizeMsg := &SetChunkSizeMessage{
		ChunkSize: setChunkSize.(*SetChunkSizeMessage).ChunkSize,
	}

	if err := h.serializeAndSendMessage(defaultChunkStreamId, setChunkSizeMsg); err != nil {
		return err
	}

	const responseChunkStreamId = 3

	if err := h.serializeAndSendMessage(responseChunkStreamId, connectSuccessResponse(h.currentTxId)); err != nil {
		return err
	} else {
		h.currentTxId += 1
	}

	if err := h.serializeAndSendMessage(responseChunkStreamId, onBwDoneResponse()); err != nil {
		return err
	}

	releaseStream, err := h.readAndParseMessage()
	if err != nil {
		return err
	}

	if releaseStream, ok := releaseStream.(*ReleaseStreamCommand); ok {
		if err := h.sendDefaultResponse(responseChunkStreamId, releaseStream.TxId, []interface{}{}); err != nil {
			return err
		}
	} else {
		return errors.New("expected a ReleaseStream command")
	}

	fcPublish, err := h.readAndParseMessage()
	if err != nil {
		return err
	}
	if _, ok := fcPublish.(*FCPublishCommand); ok {
		response := fcPublishResponse()
		if err := h.serializeAndSendMessage(responseChunkStreamId, response); err != nil {
			return err
		}
	} else {
		return errors.New("expected FCPublish command")
	}

	createStream, err := h.readAndParseMessage()
	if err != nil {
		return err
	}
	if createStream, ok := createStream.(*CreateStreamCommand); ok {
		if err := h.sendDefaultResponse(responseChunkStreamId, createStream.TxId, []interface{}{float64(1.0)}); err != nil {
			return err
		}
	} else {
		return errors.New("expected a CreateStream command")
	}

	return nil
}

func (h *handler) handleAuthorization() error {
	publish, err := h.readAndParseMessage()
	if err != nil {
		return err
	}

	if publish, ok := publish.(*PublishCommand); ok {
		if h.callbacks.OnAuthorize(publish.StreamKey) {
			const responseChunkStreamId = 3
			response := publishSuccessResponse(publish.StreamKey)
			if err := h.serializeAndSendMessage(responseChunkStreamId, response); err != nil {
				return err
			}

			h.mediaChannel <- MediaStreamInfo{StreamKey: publish.StreamKey}

			return nil
		} else {
			return errors.New("unauthorized")
		}
	} else {
		return errors.New("expected a Publish command")
	}
}

func (h *handler) handleSetDataFrame() error {
	setDataFrame, err := h.readAndParseMessage()
	if err != nil {
		return err
	}

	if setDataFrame, ok := setDataFrame.(*SetDataFrameMessage); ok {
		if h.callbacks.OnSetDataFrame(*setDataFrame) {
			return nil
		} else {
			return errors.New("setDataFrame has been rejected")
		}
	} else {
		return errors.New("expected a SetDataFrame message")
	}
}

func (h *handler) readAndParseMessage() (interface{}, error) {
	rawMsg, err := h.messageReader.ReadMessage(h.reader)
	if err != nil {
		return nil, err
	}

	return ParseMessage(rawMsg)
}

func (h *handler) handleMessage() error {
	rawMsg, err := h.messageReader.ReadMessage(h.reader)
	if err != nil {
		return err
	}

	msg, err := ParseMessage(rawMsg)
	if err != nil {
		return err
	}

	switch msg.(type) {
	case *VideoMessage:
		data := h.mediaStreamWrapper.WrapMessage(rawMsg)
		h.mediaChannel <- MediaStreamData{Data: data}
	case *AudioMessage:
		data := h.mediaStreamWrapper.WrapMessage(rawMsg)
		h.mediaChannel <- MediaStreamData{Data: data}
	}

	return nil
}

func (h *handler) serializeAndSendMessage(chunkStreamId uint8, msg MessageSerializer) error {
	msgType := msg.Type()
	msgPayload := msg.Serialize()

	payload, err := h.messageWriter.Write(&Message{
		Header: &Header{
			Type:          msgType,
			ChunkStreamId: chunkStreamId,
			Timestamp:     0,
			BodySize:      uint32(len(msgPayload)),
			StreamId:      0,
		},
		Payload: msgPayload,
	})
	if err != nil {
		return err
	}

	if _, err := h.writer.Write(payload); err != nil {
		return err
	}

	if err := h.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (h *handler) sendDefaultResponse(chunkStreamId uint8, txId float64, properties []interface{}) error {
	response := &AnonymousMessage{
		Name:       "_result",
		TxId:       &txId,
		Properties: properties,
	}

	return h.serializeAndSendMessage(chunkStreamId, response)
}

func connectSuccessResponse(txId uint32) *AnonymousMessage {
	id := float64(txId)
	return &AnonymousMessage{
		Name: "_result",
		TxId: &id,
		Properties: []interface{}{
			map[string]interface{}{
				"fsmVer":       "FMS/3,0,1,123",
				"capabilities": float64(31.0),
			},
			map[string]interface{}{
				"level":          "status",
				"code":           "NetConnection.Connect.Success",
				"descritpion":    "Connection succeeded.",
				"objectEncoding": float64(0.0),
			},
		},
	}
}

func publishSuccessResponse(streamKey string) *AnonymousMessage {
	id := float64(0.0)
	return &AnonymousMessage{
		Name: "onStatus",
		TxId: &id,
		Properties: []interface{}{
			nil,
			map[string]interface{}{
				"level":       "status",
				"code":        "NetStream.Publish.Start",
				"descritpion": fmt.Sprintf("%s is published", streamKey),
				"details":     streamKey,
			},
		},
	}
}

func onBwDoneResponse() *AnonymousMessage {
	id := float64(0.0)
	return &AnonymousMessage{
		Name: "onStatus",
		TxId: &id,
		Properties: []interface{}{
			nil,
			float64(8192.0),
		},
	}
}

func fcPublishResponse() *AnonymousMessage {
	return &AnonymousMessage{
		Name:       "onFcPublish",
		TxId:       nil,
		Properties: []interface{}{},
	}
}
