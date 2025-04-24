package oxf

import (
	"context"

	"aidicti.top/api/protogen_cmn"
	"aidicti.top/api/protogen_oxf"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/gatewaydialog/processor"
	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/model"
	"aidicti.top/pkg/model/oxf"
	"aidicti.top/pkg/model/oxf/conv"
	"aidicti.top/pkg/utils"
)

type Req struct {
	model.ReqData
	oxf.Word
}

type Resp struct {
	model.ReqData
	oxf.DictionaryEntry
	//TODO probably we should err here to check &!
}

type gateway struct {
	client protogen_oxf.ServiceOXFClient
	reqs   chan Req
	resps  chan Resp
}

func (g gateway) Reqs() chan<- Req {
	return g.reqs
}

func (g gateway) Resps() <-chan Resp {
	return g.resps
}

func New(client protogen_oxf.ServiceOXFClient) *gateway {
	return &gateway{
		client: client,
		reqs:   make(chan Req, 32),
		resps:  make(chan Resp, 32),
	}
}

type reqProcessor struct {
	clnt  protogen_oxf.ServiceOXFClient
	resps chan<- Resp
}

func (p reqProcessor) Process(ctx context.Context, req Req) error {
	utils.Assert(len(req.Text) != 0, "req word size == 0")

	resp, err := p.clnt.GetDictEntry(ctx, &protogen_oxf.Word{
		Id: &protogen_cmn.ReqData{
			ReqId: uint64(req.ReqId),
			DlgId: uint64(req.DlgId),
		},
		Text: req.Text,
	})

	//TODO rewrite errors
	if err != nil {
		logging.Warn("parse html fail", "err", err)
		return err
	}

	m := conv.FromProto(resp)
	//TODO rewrite errors
	if m == nil {
		logging.Warn("parse html fail", "err", err)
		return err
	}

	select {
	case p.resps <- Resp{ReqData: req.ReqData, DictionaryEntry: *m}:
		logging.Info("parse html fail", "err", err)

	default:
		logging.Debug("parse html fail", "err", err)
	}

	return nil
}

func (g gateway) Run(ctx context.Context) {
	reqLstn := listener.New(
		g.reqs,
		processor.New(reqProcessor{clnt: g.client, resps: g.resps}, 16))

	reqLstn.Run(ctx)
}
