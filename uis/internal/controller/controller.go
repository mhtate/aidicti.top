package controller

import (
	"context"

	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/uis/internal/model"
)

type service interface {
	CreateButton(context.Context, pkgmodel.ReqData, model.Button) (
		model.ButtonInfo, error)
	GetButton(context.Context, pkgmodel.ReqData, model.ButtonId) (
		model.ButtonInfo, error)
	ClickButton(context.Context, pkgmodel.ReqData, model.ButtonId) (
		model.ButtonInfo, error)
}

type ctrl struct {
	s service
}

func New(s service) *ctrl {
	return &ctrl{s: s}
}

func (c ctrl) CreateButton(ctx context.Context, id pkgmodel.ReqData, btn *model.Button) (
	*model.ButtonInfo, error) {

	//TODO there is a place to check validaty data
	// if btn.

	o, err := c.s.CreateButton(ctx, id, *btn)

	return &o, err

}

func (c ctrl) GetButton(ctx context.Context, id pkgmodel.ReqData, btnId *model.ButtonId) (
	*model.ButtonInfo, error) {

	o, err := c.s.GetButton(ctx, id, *btnId)

	return &o, err
}

func (c ctrl) ClickButton(ctx context.Context, id pkgmodel.ReqData, btnId *model.ButtonId) (
	*model.ButtonInfo, error) {

	o, err := c.s.ClickButton(ctx, id, *btnId)

	return &o, err
}
