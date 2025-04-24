package uis

import (
	"context"

	"aidicti.top/api/protogen_cmn"
	"aidicti.top/api/protogen_uis"
	"aidicti.top/pkg/gatewaydialog"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/scn/internal/model"
)

type Req struct {
	pkgmodel.ReqData
	model.Button
}

type Resp struct {
	pkgmodel.ReqData
	model.ButtonId
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
	btn := model.FromButton(req.Button)

	btn.Id = &protogen_cmn.ReqData{ReqId: uint64(req.ReqId), DlgId: uint64(req.DlgId)}

	resp, err := p.clnt.CreateButton(ctx, btn)

	//TODO rewrite errors
	if err != nil {
		logging.Warn("create button fail", "err", err)
		return err
	}

	//TODO rewrite errors
	// if m == nil {
	// 	logging.Warn("parse html fail", "err", err)
	// 	return err
	// }

	select {
	case p.resps <- Resp{
		ReqData:  req.ReqData,
		ButtonId: model.ButtonId(resp.ButtonId.ButtonId)}:

		logging.Info("create button ok", "err", err)

	default:
		logging.Debug("create button fail", "err", err)
	}

	return nil
}

func (g gateway) Run(ctx context.Context) {
	// utils.Assert(g.initPrF != nil, "req word size == 0")

	// reqLstn := listener.New(g.reqs, g.initPrF(reqProcessor{g.resps, g.clnt}))

	reqLstn := listener.New(g.reqs, reqProcessor{g.resps, g.clnt})

	reqLstn.Run(ctx)
}
