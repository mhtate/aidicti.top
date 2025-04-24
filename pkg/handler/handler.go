package handler

import (
	"context"
)

type Handler[Req any, Resp any] interface {
	Exec(context.Context, Req) (Resp, error)
}
