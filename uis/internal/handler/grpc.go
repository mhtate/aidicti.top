package handler

import (
	"context"

	"aidicti.top/api/protogen_cmn"
	"aidicti.top/api/protogen_uis"
	"aidicti.top/pkg/handler/exectime"
	"aidicti.top/pkg/handler/proc"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/uis/internal/model"
)

//TODO rename all entries
// protogen "aidicti.top/api/protogen_uis"

type controller interface {
	CreateButton(context.Context, pkgmodel.ReqData, *model.Button) (
		*model.ButtonInfo, error)
	GetButton(context.Context, pkgmodel.ReqData, *model.ButtonId) (
		*model.ButtonInfo, error)
	ClickButton(context.Context, pkgmodel.ReqData, *model.ButtonId) (
		*model.ButtonInfo, error)
}

type handlerProc_CreateButton interface {
	Exec(context.Context, *protogen_uis.Button) (
		*protogen_uis.ButtonInfo, error)
}

type handlerProc_GetButton interface {
	Exec(context.Context, *protogen_uis.ButtonId) (
		*protogen_uis.ButtonInfo, error)
}

type handlerProc_ClickButton interface {
	Exec(context.Context, *protogen_uis.ButtonId) (
		*protogen_uis.ButtonInfo, error)
}

type handler struct {
	protogen_uis.UnimplementedServiceUISServer
	ctrl controller

	procCreateButton handlerProc_CreateButton
	procGetButton    handlerProc_GetButton
	procClickButton  handlerProc_ClickButton
}

func New(c controller) *handler {
	h := &handler{ctrl: c}

	h.procCreateButton = exectime.New(
		proc.New(
			func(req *protogen_uis.Button) (*model.Button, error) {
				btn := model.ToButton(req)
				return &btn, nil
			},

			func(resp *model.ButtonInfo) (*protogen_uis.ButtonInfo, error) {
				info := model.FromButtonInfo(*resp)

				info.Id = &protogen_cmn.ReqData{}

				return info, nil
			},

			h.ctrl.CreateButton,
		))

	h.procGetButton = exectime.New(
		proc.New(
			func(req *protogen_uis.ButtonId) (*model.ButtonId, error) {
				btnid := model.ButtonId(req.ButtonId)
				return &btnid, nil
			},

			func(resp *model.ButtonInfo) (*protogen_uis.ButtonInfo, error) {
				info := model.FromButtonInfo(*resp)

				info.Id = &protogen_cmn.ReqData{}

				return info, nil
			},

			h.ctrl.GetButton,
		))

	h.procClickButton = exectime.New(
		proc.New(
			func(req *protogen_uis.ButtonId) (*model.ButtonId, error) {
				btnid := model.ButtonId(req.ButtonId)
				return &btnid, nil
			},

			func(resp *model.ButtonInfo) (*protogen_uis.ButtonInfo, error) {
				info := model.FromButtonInfo(*resp)

				info.Id = &protogen_cmn.ReqData{}

				return info, nil
			},

			h.ctrl.ClickButton,
		))

	return h
}

func (h handler) CreateButton(
	ctx context.Context,
	btn *protogen_uis.Button) (*protogen_uis.ButtonInfo, error) {

	return h.procCreateButton.Exec(ctx, btn)
}

func (h handler) GetButton(
	ctx context.Context,
	id *protogen_uis.ButtonId) (*protogen_uis.ButtonInfo, error) {

	return h.procGetButton.Exec(ctx, id)
}

func (h handler) ClickButton(
	ctx context.Context,
	id *protogen_uis.ButtonId) (*protogen_uis.ButtonInfo, error) {

	return h.procClickButton.Exec(ctx, id)
}
