package listener

import (
	"context"

	"aidicti.top/pkg/gatewaydialog"
	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/utils"
)

type listener[T any] struct {
	ch   <-chan T
	pr   gatewaydialog.Processor[T]
	name string
}

func New[T any](ch <-chan T, pr gatewaydialog.Processor[T]) *listener[T] {
	l := &listener[T]{ch: ch, pr: pr}
	l.name = utils.GetTypeName(*l)

	return l
}

func (l *listener[T]) Run(ctx context.Context) {
	for {
		select {
		case req := <-l.ch:
			logging.Debug("process req", "st", "start", "type", l.name)

			err := l.pr.Process(ctx, req)
			if err != nil {
				logging.Warn("process req", "st", "fail", "err", err, "type", l.name)
			}

		case <-ctx.Done():
			logging.Info("listen req", "st", "finish", "rsn", "ctx done", "type", l.name)
			return
		}
	}
}
