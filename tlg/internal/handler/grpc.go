package handler

import (
	"context"
	"sync"

	"aidicti.top/api/protogen_tlg"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/tlg/internal/model"
	"google.golang.org/grpc"
)

type controller interface {
	//TODO change the name of a method
	GetMessage(reqData pkgmodel.ReqData, message model.Message)
	SendMessage(message model.Message) error
}

type handler struct {
	ctrl controller
	clnt protogen_tlg.TelegramServiceClient
	strm stream
	ctx  context.Context
}

type client protogen_tlg.TelegramServiceClient
type stream grpc.BidiStreamingClient[protogen_tlg.Message, protogen_tlg.Message]

func New(ctx context.Context, c controller, clnt protogen_tlg.TelegramServiceClient) *handler {
	h := &handler{ctx: ctx, ctrl: c, clnt: clnt}

	return h
}

func (h *handler) getStream() (stream, error) {
	if h.strm != nil {
		return h.strm, nil
	}

	//TODO check we fail to recconet
	stream, err := h.clnt.TelegramChat(h.ctx)
	if err != nil {
		logging.Warn("create stream fail", "err", err)
		return nil, err
	}

	logging.Debug("create stream ok")
	h.strm = stream

	wg := sync.WaitGroup{}
	wg.Add(1)

	listenToStream := func() {
		wg.Done()
		for {
			resp, err := h.strm.Recv()
			if err != nil {
				logging.Info("close stream goro", "err", err)

				h.strm = nil

				return
			}

			logging.Debug("receive msg in stream goro ok")

			h.ReceiveMessage(resp)
		}
	}

	go listenToStream()

	wg.Wait()

	return h.strm, nil
}

func (h *handler) ReceiveMessage(message *protogen_tlg.Message) error {
	logging.Info("c")

	m := model.Message{
		Id:      model.UserID(message.UserId),
		Message: message.Message,
		Actions: message.Actions,
	}

	if message.Audio != nil {
		m.Audio = &model.AudioData{Data: message.Audio.Data}
	}

	if message.Action != nil {
		m.Action = &model.Action{
			Type:    message.Action.Type,
			Message: message.Action.Message,
			Values:  message.Action.Values,
		}
	}

	logging.Info("b")

	h.ctrl.GetMessage(pkgmodel.ReqData{DlgId: pkgmodel.DlgID(message.UserId)}, m)

	//TODO what is it
	return nil
}

func (h *handler) SendMessage(message model.Message) error {
	m := &protogen_tlg.Message{
		UserId:  uint64(message.Id),
		Message: message.Message,
	}

	if message.Audio != nil {
		m.Audio = &protogen_tlg.AudioFile{Data: message.Audio.Data}
	}

	if message.Action != nil {
		m.Action = &protogen_tlg.Action{
			Type:    message.Action.Type,
			Message: message.Action.Message,
			Values:  message.Action.Values,
		}
	}

	stream, err := h.getStream()
	if err != nil {
		logging.Warn("get stream fail", "err", err)
		return err
	}

	err = stream.Send(m)
	if err != nil {
		logging.Warn("send to stream fail", "err", err)
		return err
	}

	logging.Debug("send to stream ok")
	return nil
}
