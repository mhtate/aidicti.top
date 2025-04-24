package youheardright

import (
	"encoding/json"

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
	Txt    string
}

type meta struct {
	Txt string `json:"txt"`
}

// TODO we return error why here we decided to return nil &!
func New(btn model.Button, scn gatewaySCN) *button {
	mt := meta{}
	err := json.Unmarshal(btn.Meta, &mt)
	if err != nil {
		logging.Warn("creating btn is error")
		return nil
	}

	return &button{
		text: btn.Texts,
		Txt:  mt.Txt,
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
		Message: model.Message{Message: b.Txt},
	}
}
