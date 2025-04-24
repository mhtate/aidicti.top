package handler

import (
	"net"
	"sync"

	"aidicti.top/api/protogen_tlg"
	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/utils"
	"aidicti.top/scn/internal/model"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type controller interface {
	In(model.Message)
	Out() <-chan model.Message
}

func fromProto(message *protogen_tlg.Message) model.Message {
	m := model.Message{Id: model.UserID(message.UserId), Message: message.Message}

	if message.Audio != nil {
		if len(message.Audio.Data) != 0 {
			m.Audio = &model.AudioData{Data: message.Audio.Data}
		}
	}

	if message.Action != nil {
		utils.Assert(m.Action.Type != "", "")
		utils.Assert(m.Action.Message != "", "")
		utils.Assert(len(message.Action.Values) != 0, "")

		m.Action = &model.Action{
			Type:    m.Action.Type,
			Message: m.Action.Message,
			Values:  message.Action.Values,
		}
	}

	logging.Debug("message from proto", "msg", m)
	return m
}

func toProto(message model.Message) *protogen_tlg.Message {
	//TODO check id
	m := &protogen_tlg.Message{
		UserId:  uint64(message.Id),
		Message: message.Message,
		Actions: message.Actions}

	if message.Audio != nil {
		utils.Assert(len(message.Audio.Data) != 0, "model audio has 0 len")

		m.Audio = &protogen_tlg.AudioFile{Data: message.Audio.Data}
	}

	if message.Action != nil {
		utils.Assert(message.Action.Type != "", "")
		utils.Assert(message.Action.Message != "", "")
		utils.Assert(len(message.Action.Values) != 0, "")

		m.Action = &protogen_tlg.Action{
			Type:    message.Action.Type,
			Message: message.Action.Message,
			Values:  message.Action.Values}
	}

	logging.Debug("message to proto", "msg", m)
	return m
}

type stream grpc.BidiStreamingServer[protogen_tlg.Message, protogen_tlg.Message]

type chatServer struct {
	protogen_tlg.UnimplementedTelegramServiceServer

	ctrl    controller
	streams map[uuid.UUID]stream
	mu      sync.Mutex
}

func New(c controller) *chatServer {
	s := &chatServer{ctrl: c, streams: make(map[uuid.UUID]stream)}

	notifyStreams := func(m *protogen_tlg.Message) {
		s.mu.Lock()
		defer s.mu.Unlock()

		for _, stream := range s.streams {
			err := stream.Send(m)
			if err != nil {
				logging.Info("stream sending error", "err", err)
			}
		}

	}

	go func() {
		for {
			select {
			case m, ok := <-c.Out():
				if !ok {
					logging.Info("ctrl closed chan")
					return
				}

				notifyStreams(toProto(m))

				logging.Debug("got message from ctrl", "msg", m)
			}
		}
	}()

	return s
}

func (s *chatServer) TelegramChat(stream grpc.BidiStreamingServer[protogen_tlg.Message, protogen_tlg.Message]) error {
	id, err := uuid.NewUUID()
	utils.Assert(err == nil, "UUID creating problem")

	safeInsert := func() {
		s.mu.Lock()
		logging.Info("creating stream", "id", id)
		s.streams[id] = stream
		s.mu.Unlock()
	}

	safeInsert()

	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		logging.Info("deleting stream", "id", id)
		delete(s.streams, id)
	}()

	p, ok := peer.FromContext(stream.Context())
	if !ok {
		logging.Info("Could not get peer info")
	} else {
		logging.Info("Client address", "addr", p.Addr)
		if tcpAddr, ok := p.Addr.(*net.TCPAddr); ok {
			logging.Info("Client port", "port", tcpAddr.Port)
		}
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			logging.Info("stream error", err)
			return err
		}

		m := fromProto(req)

		logging.Debug("passing to controller", "message", m)
		s.ctrl.In(m)
	}
}
