package storage

import (
	"context"
	"fmt"
	"sync"

	"aidicti.top/pkg/gatewaydialog"
	"aidicti.top/pkg/model"
)

type dialogStorage[Req any, Resp any] struct {
	mp           map[model.DlgID]gatewaydialog.Dialog[Req, Resp]
	mtx          sync.Mutex
	createDialog func(id model.DlgID) (gatewaydialog.Dialog[Req, Resp], error)
	ctx          context.Context
}

func New[Req any, Resp any](
	ctx context.Context, createDialog func(id model.DlgID) (gatewaydialog.Dialog[Req, Resp], error)) *dialogStorage[Req, Resp] {

	return &dialogStorage[Req, Resp]{
		mp:           make(map[model.DlgID]gatewaydialog.Dialog[Req, Resp]),
		ctx:          ctx,
		createDialog: createDialog}
}

func (s *dialogStorage[Req, Resp]) Run(ctx context.Context) {
	//TODO move all here and dont put ctx in New
	s.ctx = ctx
}

func (s *dialogStorage[Req, Resp]) Get(id model.DlgID) (gatewaydialog.Dialog[Req, Resp], error) {
	select {
	case <-s.ctx.Done():
		return nil, fmt.Errorf("ctx done")
	default:
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	d, exists := s.mp[id]

	if exists {
		return d, nil
	}

	dlg, err := s.createDialog(id)
	if err != nil {
		return nil, err
	}

	go dlg.Run(s.ctx)

	s.mp[id] = dlg

	return dlg, nil
}
