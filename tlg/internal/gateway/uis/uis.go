package uis

import (
	"context"
	"time"

	"aidicti.top/api/protogen_cmn"
	"aidicti.top/api/protogen_uis"
	"aidicti.top/pkg/gatewaydialog"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/tlg/internal/model"
)

type Req struct {
	pkgmodel.ReqData
	model.ButtonId
	Clicked bool
}

type Resp struct {
	pkgmodel.ReqData
	model.ButtonInfo
}

type gateway struct {
	reqs  chan Req
	resps chan Resp
	clnt  protogen_uis.ServiceUISClient
	// initPrF func(gatewaydialog.Processor[Req]) gatewaydialog.Processor[Req]
}

func (g gateway) Reqs() chan<- Req {
	return g.reqs
}

func (g gateway) Resps() <-chan Resp {
	return g.resps
}

func (g *gateway) SetProcessor(
	pr func(gatewaydialog.Processor[Req]) gatewaydialog.Processor[Req]) {
	// g.initPrF = pr
}

func New(clnt protogen_uis.ServiceUISClient) *gateway {
	return &gateway{
		reqs:  make(chan Req, 64),
		resps: make(chan Resp, 64),
		clnt:  clnt,
	}
}

type reqProcessor struct {
	resps chan<- Resp
	clnt  protogen_uis.ServiceUISClient
}

func NewProcessor() *reqProcessor {
	return &reqProcessor{}
}

func (p reqProcessor) Process(ctx context.Context, req Req) error {
	var resp *protogen_uis.ButtonInfo
	var err error

	if req.Clicked {
		resp, err = p.clnt.ClickButton(ctx, &protogen_uis.ButtonId{
			Id:       &protogen_cmn.ReqData{ReqId: uint64(req.ReqData.ReqId), DlgId: uint64(req.DlgId)},
			ButtonId: uint64(req.ButtonId)})
		if err != nil {
			//TODO
			return err
		}
	} else {
		resp, err = p.clnt.GetButton(ctx, &protogen_uis.ButtonId{
			Id:       &protogen_cmn.ReqData{ReqId: uint64(req.ReqData.ReqId), DlgId: uint64(req.DlgId)},
			ButtonId: uint64(req.ButtonId)})
		if err != nil {
			//TODO
			return err
		}
	}

	select {
	case p.resps <- Resp{ReqData: req.ReqData, ButtonInfo: model.ToButtonInfo(resp)}:
		logging.Info("create button ok", "err", err)

	case <-time.After(5 * time.Second):
		logging.Debug("create button fail [timeout]", "err", err)

	default:
		logging.Debug("create button fail  [done]", "err", err)
	}

	return nil
}

func (g gateway) Run(ctx context.Context) {
	// utils.Assert(g.initPrF != nil, "req word size == 0")

	// reqLstn := listener.New(g.reqs, g.initPrF(reqProcessor{g.resps, g.clnt}))

	reqLstn := listener.New(g.reqs, reqProcessor{g.resps, g.clnt})

	reqLstn.Run(ctx)
}
