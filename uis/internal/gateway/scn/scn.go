package scn

import (
	"context"

	"aidicti.top/api/protogen_tlg"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/gatewaydialog/processor"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/uis/internal/model"
)

type Req struct {
	pkgmodel.ReqData
	model.Message
}

type Resp struct{}

type gateway struct {
	clnt  protogen_tlg.TelegramServiceClient
	reqs  chan Req
	resps chan Resp
}

func (g gateway) Reqs() chan<- Req {
	return g.reqs
}

func (g gateway) Resps() <-chan Resp {
	return g.resps
}

func New(clnt protogen_tlg.TelegramServiceClient) *gateway {
	return &gateway{
		clnt:  clnt,
		reqs:  make(chan Req, 32),
		resps: make(chan Resp, 32),
	}
}

type reqProcessor struct {
	resps chan<- Resp
	clnt  protogen_tlg.TelegramServiceClient
}

func (p reqProcessor) Process(ctx context.Context, req Req) error {
	stream, err := p.clnt.TelegramChat(ctx)
	if err != nil {
		logging.Warn("create scn stream", "st", "fail", "err", err)
		return err
	}

	err = stream.Send(&protogen_tlg.Message{
		UserId:  uint64(req.DlgId),
		Message: req.Message.Message,
	})

	if err != nil {
		logging.Warn("send scn stream", "st", "fail", "err", err)
	}

	err = stream.CloseSend()
	if err != nil {
		logging.Warn("close scn stream", "st", "fail", "err", err)
	}

	return nil
}

func (g gateway) Run(ctx context.Context) {
	reqLstn := listener.New(
		g.reqs,
		processor.New(reqProcessor{resps: g.resps, clnt: g.clnt}, 16))

	reqLstn.Run(ctx)
}
