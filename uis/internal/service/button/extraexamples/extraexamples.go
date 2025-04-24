package extraexamples

import (
	"encoding/json"

	"aidicti.top/pkg/logging"
	"aidicti.top/uis/internal/gateway/dbs"
	"aidicti.top/uis/internal/model"
)

// type gatewayDBS = gatewaydialog.Dialog[dbs.Req, dbs.Resp]

type gatewayDBS interface {
	Reqs() chan<- dbs.Req
	Resps() <-chan dbs.Resp
}

type button struct {
	text []string
	mt   meta
	//TODO KOSTYL KOSTYL
	UserId  uint64
	exposed bool
}

type meta struct {
	Examples string `json:"examples"`
}

// TODO we return error why here we decided to return nil &!
func New(btn model.Button) *button {
	mt := meta{}
	err := json.Unmarshal(btn.Meta, &mt)
	if err != nil {
		logging.Warn("creating btn is error")
		return nil
	}

	return &button{
		text:    btn.Texts,
		mt:      mt,
		exposed: false,
	}
}

func (b *button) Info() model.ButtonInfo {
	info := model.ButtonInfo{}

	if b.exposed {
		info.Text = ""
	}

	return info
}

func (b *button) Click() {
	b.exposed = true
}
