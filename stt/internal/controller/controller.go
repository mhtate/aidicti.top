package controller

import (
	"context"

	"aidicti.top/stt/internal/model"

	pkgmodel "aidicti.top/pkg/model"
)

type service interface {
	TranslateAudio(context.Context, model.RequestID, model.UserID, model.AudioData) (
		model.TranscriptionResult, error)
}

type controller struct {
	s service
}

func New(s service) *controller {
	return &controller{s: s}
}

func (c controller) TranslateAudio(ctx context.Context, id pkgmodel.ReqData, data *model.AudioData) (
	*model.TranscriptionResult, error) {

	res, err := c.s.TranslateAudio(ctx, model.RequestID(id.ReqId), model.UserID(id.DlgId), *data)

	return &res, err
}
