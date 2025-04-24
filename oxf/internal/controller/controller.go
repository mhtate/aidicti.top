package controller

import (
	"context"

	"aidicti.top/oxf/internal/model"
	pkgmodel "aidicti.top/pkg/model"
)

type service interface {
	GetDictEntry(ctx context.Context, ids pkgmodel.ReqData, word model.Word) (
		model.DictionaryEntry, error)
}

type ctrl struct {
	s service
}

func New(s service) *ctrl {
	return &ctrl{s: s}
}

func (c ctrl) GetDictEntry(
	ctx context.Context,
	ids pkgmodel.ReqData,
	word *model.Word) (*model.DictionaryEntry, error) {

	dict, err := c.s.GetDictEntry(ctx, ids, *word)

	return &dict, err
}
