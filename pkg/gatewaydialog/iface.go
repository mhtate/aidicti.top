package gatewaydialog

import (
	"context"

	"aidicti.top/pkg/model"
)

type Gateway[GtwReq any, GtwResp any] interface {
	Reqs() chan<- GtwReq
	Resps() <-chan GtwResp
}

type Dialog[Req any, Resp any] interface {
	Run(context.Context)
	Reqs() chan Req
	Resps() chan Resp
}

type Storage[Req any, Resp any] interface {
	Get(model.DlgID) (Dialog[Req, Resp], error)
}

type Processor[Req any] interface {
	Process(context.Context, Req) error
}
