package controller

import (
	"context"

	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/tlg/internal/model"
)

type service interface {
	GetMessage(context.Context, pkgmodel.ReqData, model.Message)
}

type handler interface {
	SendMessage(model.Message) error
}

type controller struct {
	s service
	H handler
}

func New(s service, h handler) *controller {
	return &controller{s: s, H: h}
}

func (c *controller) GetMessage(reqData pkgmodel.ReqData, message model.Message) {
	logging.Info("c")
	c.s.GetMessage(context.Background(), reqData, message)
}

func (c *controller) SendMessage(message model.Message) error {
	logging.Info("s")
	return c.H.SendMessage(message)
}
