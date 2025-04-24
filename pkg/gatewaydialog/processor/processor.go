package processor

import (
	"context"

	"aidicti.top/pkg/gatewaydialog"
	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/utils"
)

type BatchCount uint8

type batchedProcessor[T any] struct {
	pr  gatewaydialog.Processor[T]
	sem chan struct{}
}

func New[T any](pr gatewaydialog.Processor[T], batches BatchCount) *batchedProcessor[T] {
	utils.Assert(batches != 0, "batches == 0")

	return &batchedProcessor[T]{pr: pr, sem: make(chan struct{}, batches)}
}

func (p batchedProcessor[T]) Process(ctx context.Context, t T) error {
	select {
	case p.sem <- struct{}{}:
		go func(t T) {
			defer func() { <-p.sem }()

			err := p.pr.Process(ctx, t)
			if err != nil {
				logging.Debug("process req", "st", "fail")
			}
		}(t)

	case <-ctx.Done():
		return nil
	}

	return nil
}
