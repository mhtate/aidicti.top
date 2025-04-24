package exectime

import (
	"context"
	"fmt"
	"time"

	"aidicti.top/pkg/handler"
	"aidicti.top/pkg/logging"
)

type handlerExec[Req any, Resp any] struct {
	hnd handler.Handler[Req, Resp]
}

func New[Req any, Resp any](hnd handler.Handler[Req, Resp]) *handlerExec[Req, Resp] {
	return &handlerExec[Req, Resp]{hnd: hnd}
}

func (p handlerExec[Req, Resp]) Exec(ctx context.Context, req Req) (Resp, error) {
	start := time.Now()

	resp, err := p.hnd.Exec(ctx, req)

	logging.Info("proc req", "time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
	logging.Info("proc req", "time", fmt.Sprintf("%dns", time.Since(start).Nanoseconds()))

	return resp, err
}
