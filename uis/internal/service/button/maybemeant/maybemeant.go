package maybemeant

import (
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/uis/internal/gateway/scn"
	"aidicti.top/uis/internal/model"
)

type gatewaySCN interface {
	Reqs() chan<- scn.Req
	Resps() <-chan scn.Resp
}

type button struct {
	text   []string
	scn    gatewaySCN
	UserId uint64
}

// TODO we return error why here we decided to return nil &!
func New(btn model.Button, scn gatewaySCN) *button {
	return &button{
		text: btn.Texts,
		scn:  scn,
	}
}

func (b *button) Info() model.ButtonInfo {
	info := model.ButtonInfo{}
	info.Text = b.text[0]

	return info
}

func (b *button) Click() {
	logging.Info("button clicked")

	b.scn.Reqs() <- scn.Req{
		ReqData: pkgmodel.ReqData{DlgId: pkgmodel.DlgID(b.UserId)},
		Message: model.Message{Message: b.text[0]},
	}
}
