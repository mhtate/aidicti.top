package retry

import (
	"context"
	"time"

	"aidicti.top/pkg/gatewaydialog"
	"aidicti.top/pkg/logging"
	"github.com/slok/goresilience"
	"github.com/slok/goresilience/retry"
)

type retryProcessor[Req any] struct {
	pr    gatewaydialog.Processor[Req]
	retry goresilience.Runner
}

func New[T any](pr gatewaydialog.Processor[T], WaitBase time.Duration, Backoff bool, Times int) *retryProcessor[T] {
	retry := retry.New(retry.Config{WaitBase: WaitBase, DisableBackoff: !Backoff, Times: Times})

	return &retryProcessor[T]{pr: pr, retry: retry}
}

func (p retryProcessor[T]) Process(ctx context.Context, t T) error {
	retryCount := 0

	return p.retry.Run(ctx, func(ctx context.Context) error {
		logging.Debug("process req", "st", "start", "retry", retryCount)
		retryCount += 1

		err := p.pr.Process(ctx, t)
		if err != nil {
			logging.Debug("process req", "st", "fail", "retry", retryCount, "err", err)
			return err
		}

		return nil
	})
}
