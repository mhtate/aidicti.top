package controller

import (
	"context"

	"aidicti.top/pkg/logging"
	"aidicti.top/scn/internal/model"
)

type controller struct {
	requests  chan model.Message
	responses chan model.Message
}

func New() *controller {
	return &controller{requests: make(chan model.Message, 256), responses: make(chan model.Message, 256)}
}

func (c controller) In(m model.Message) {
	select {
	case c.responses <- m:
		logging.Debug("ctrl responses put", "msg", m)

	default:
		logging.Info("ctrl responses chan full")
	}
}

func (c controller) Out() <-chan model.Message {
	return c.requests
}

func (c controller) Requests() chan<- model.Message {
	return c.requests
}

func (c controller) Responses() <-chan model.Message {
	return c.responses
}

func (c controller) Reqs() chan<- model.Message {
	return c.requests
}

func (c controller) Resps() <-chan model.Message {
	return c.responses
}

func (c controller) Run(ctx context.Context) {
	select {
	case <-ctx.Done():

	}
}
