package model

import (
	"aidicti.top/api/protogen_uis"
	"aidicti.top/pkg/utils"
)

type Button struct {
	Tp    ButtonType
	Texts []string
	Meta  []byte
}

type ButtonInfo struct {
	Id   ButtonId
	Text string
}

type ButtonId uint64

type ButtonType uint

const (
	AddToDictButton     = ButtonType(1)
	ExtraExamplesButton = ButtonType(2)
	MayBeMeantButton    = ButtonType(3)
	YouHeardRight       = ButtonType(4)
)

// TODO we should do smth with that
func ToButton(btn *protogen_uis.Button) Button {
	utils.Assert(btn != nil, "")

	out := Button{}

	out.Tp = ButtonType(btn.Type)

	out.Texts = btn.Text

	out.Meta = btn.Meta

	return out
}

// TODO get known where we use pointer where value of model
func FromButton(btn Button) *protogen_uis.Button {
	// utils.Assert(btn != nil, "")

	out := &protogen_uis.Button{}

	out.Type = uint64(btn.Tp)

	out.Text = btn.Texts

	out.Meta = btn.Meta

	return out
}

func ToButtonInfo(btn *protogen_uis.ButtonInfo) ButtonInfo {
	utils.Assert(btn != nil, "")

	out := ButtonInfo{}

	out.Id = ButtonId(btn.ButtonId.ButtonId)

	out.Text = btn.Text

	return out
}

func FromButtonInfo(btn ButtonInfo) *protogen_uis.ButtonInfo {
	// utils.Assert(btn != nil, "")

	out := &protogen_uis.ButtonInfo{}

	out.ButtonId = &protogen_uis.ButtonId{ButtonId: uint64(btn.Id)}

	out.Text = btn.Text

	return out
}
