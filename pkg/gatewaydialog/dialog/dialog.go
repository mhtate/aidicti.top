package dialog

import (
	"context"

	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
)

// TODO dialog should name himself in logs
type gatewayChan[GtwReq any] interface {
	Reqs() chan<- GtwReq
}

type ReqConverter[Req any, GtwReq any] interface {
	Convert(req Req) (GtwReq, error)
}

type dialog[Req any, Resp any, GtwReq any] struct {
	id      model.DlgID
	gtw     gatewayChan[GtwReq]
	reqs    chan Req
	resps   chan Resp
	reqConv ReqConverter[Req, GtwReq]
	name    string
}

func New[Resp any, Req any, GtwReq any](
	id model.DlgID,
	gtw gatewayChan[GtwReq],
	reqConverter ReqConverter[Req, GtwReq]) *dialog[Req, Resp, GtwReq] {

	d := &dialog[Req, Resp, GtwReq]{
		id:      id,
		gtw:     gtw,
		reqs:    make(chan Req, 1),
		resps:   make(chan Resp, 1),
		reqConv: reqConverter,
	}

	d.name = utils.GetTypeName(*d)

	return d
}

func (d dialog[Req, Resp, GtwReq]) Run(ctx context.Context) {
	for {
		select {
		case req := <-d.reqs:
			gtwReq, err := d.reqConv.Convert(req)
			if err != nil {
				logging.Warn("conv req to gtw req", "st", "fail", "err", err, "type", d.name)
				break
			}

			select {
			case d.gtw.Reqs() <- gtwReq:
				logging.Debug("pass req to gtw", "st", "ok", "type", d.name)

			default:
				logging.Warn("pass req to gtw", "st", "fail", "rsn", "ch closed or full", "dlgId", d.id, "type", d.name)
			}

		case <-ctx.Done():
			logging.Info("listen resp from dlg", "st", "finished", "rsn", "ctx done", "dlgId", d.id, "type", d.name)
			return
		}
	}
}

func (d dialog[Req, Resp, GtwReq]) Resps() chan Resp {
	return d.resps
}

func (d dialog[Req, Resp, GtwReq]) Reqs() chan Req {
	return d.reqs
}
