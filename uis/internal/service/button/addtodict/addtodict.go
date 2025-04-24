package addtodict

import (
	"encoding/json"
	"strconv"
	"time"

	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/uis/internal/gateway/dbs"
	"aidicti.top/uis/internal/model"
)

// type gatewayDBS = gatewaydialog.Dialog[dbs.Req, dbs.Resp]

type gatewayDBS interface {
	Reqs() chan<- dbs.Req
	Resps() <-chan dbs.Resp
}

type executedData struct {
	Checked  bool
	Executed bool
}

type button struct {
	text []string
	dbs  gatewayDBS
	exec executedData
	mt   meta
	//TODO KOSTYL KOSTYL
	UserId uint64
}

type meta struct {
	Word  string `json:"word"`
	Def   string `json:"def"`
	Usage string `json:"usage"`
	Link  string `json:"link"`
	Pos   string `json:"pos"`
}

// TODO we return error why here we decided to return nil &!
func New(btn model.Button, dbs gatewayDBS) *button {
	mt := meta{}
	err := json.Unmarshal(btn.Meta, &mt)
	if err != nil {
		logging.Warn("creating btn is error")
		return nil
	}

	return &button{
		text: btn.Texts,
		dbs:  dbs,
		exec: executedData{false, false},
		mt:   mt,
	}
}

func (b *button) Info() model.ButtonInfo {
	info := model.ButtonInfo{}

	if !b.exec.Checked {
		b.check()
	}

	c := 0
	if b.exec.Executed {
		c = 1
	}

	info.Text = b.text[c]

	return info
}

func (b *button) Click() {

	if b.exec.Checked && b.exec.Executed {
		return
	}

	pos, _ := strconv.Atoi(b.mt.Pos)

	b.dbs.Reqs() <- dbs.Req{
		ReqData: pkgmodel.ReqData{DlgId: pkgmodel.DlgID(b.UserId)},
		DictWord: dbs.DictWord{
			Word:   b.mt.Word,
			Def:    b.mt.Def,
			Usage:  b.mt.Usage,
			Link:   b.mt.Link,
			UserId: uint(b.UserId),
			Pos:    pos,
		},
		Method: dbs.Set,
	}

	b.exec.Checked = true
	b.exec.Executed = true
}

func (b *button) check() {

	if !b.exec.Checked {
		pos, _ := strconv.Atoi(b.mt.Pos)

		b.dbs.Reqs() <- dbs.Req{
			ReqData: pkgmodel.ReqData{DlgId: pkgmodel.DlgID(b.UserId)},
			DictWord: dbs.DictWord{
				Word:   b.mt.Word,
				Def:    b.mt.Def,
				Usage:  b.mt.Usage,
				Link:   b.mt.Link,
				UserId: uint(b.UserId),
				Pos:    pos,
			},
			Method: dbs.Get,
		}

		select {
		case resp := <-b.dbs.Resps():
			logging.Info("Got answer")
			b.exec.Checked = true
			if len(resp.Words) > 0 {
				b.exec.Executed = true
			}

		case <-time.After(3 * time.Second):
			logging.Info("Got timout")
		}
	}
}
