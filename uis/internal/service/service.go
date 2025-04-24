package service

import (
	"context"
	"fmt"

	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/uis/internal/gateway/dbs"
	"aidicti.top/uis/internal/gateway/scn"
	"aidicti.top/uis/internal/model"
	"aidicti.top/uis/internal/repo"
	"aidicti.top/uis/internal/service/button"
	"aidicti.top/uis/internal/service/button/addtodict"
	"aidicti.top/uis/internal/service/button/extraexamples"
	"aidicti.top/uis/internal/service/button/maybemeant"
	"aidicti.top/uis/internal/service/button/youheardright"
)

type gatewayDBS interface {
	Reqs() chan<- dbs.Req
	Resps() <-chan dbs.Resp
}

type gatewaySCN interface {
	Reqs() chan<- scn.Req
	Resps() <-chan scn.Resp
}

type service struct {
	repo repo.Repository
	dbs  gatewayDBS
	scn  gatewaySCN
}

func New(repo repo.Repository, dbs gatewayDBS, scn gatewaySCN) *service {
	return &service{repo: repo, dbs: dbs, scn: scn}
}

// type reqConverter struct {
// 	id pkgmodel.DlgID
// }

// func (c reqConverter) Convert(req model.Word) (oxf.Req, error) {
// 	reqId := pkgmodel.ReqID(rand.Uint64())

// 	return oxf.Req{Word: req, ReqData: pkgmodel.ReqData{DlgId: c.id, ReqId: reqId}}, nil
// }

func (s *service) Run(ctx context.Context) {
	// createFunc := func(id pkgmodel.DlgID) (
	// 	gatewaydialog.Dialog[model.Word, model.DictionaryEntry], error) {

	// 	cnv := reqConverter{id: id}

	// 	logging.Debug("create dlg", "st", "ok", "id", id)

	// 	return pkgdialog.New[model.DictionaryEntry](id, s.gtw, cnv), nil
	// }

	// s.stg = pkgstorage.New(ctx, createFunc)

	// for {
	// 	select {
	// 	case resp := <-s.gtw.Resps():
	// 		logging.Debug("get resp from gtw", "st", "ok", "dlgId", resp.DlgId, "reqId", resp.ReqId)

	// 		dialog, err := s.stg.Get(resp.DlgId)
	// 		if err != nil {
	// 			logging.Warn("get dlg from stg", "st", "fail", "dlgId", resp.DlgId, "reqId", resp.ReqId, "err", err)
	// 			break
	// 		}

	// 		select {
	// 		case dialog.Resps() <- resp.DictionaryEntry:
	// 			logging.Debug("pass resp to dlg", "st", "ok", "dlgId", resp.DlgId)

	// 		default:
	// 			logging.Warn("pass resp to dlg", "st", "fail", "rsn", "ch closed or full")
	// 		}

	// 	case <-ctx.Done():
	// 		logging.Info("listen resp from gtw", "st", "finished", "rsn", "ctx done")
	// 		return
	// 	}
	// }

	select {}
}

// func (s service) GetDictEntry(
// 	ctx context.Context,
// 	ids pkgmodel.ReqData,
// 	word model.Word) (model.DictionaryEntry, error) {

// 	dialog, err := s.stg.Get(ids.DlgId)
// 	if err != nil {
// 		logging.Warn("get dlg from stg", "st", "fail", "dlgId", ids.DlgId, "reqId", ids.ReqId, "err", err)

// 		return model.DictionaryEntry{}, err
// 	}

// 	select {
// 	case dialog.Reqs() <- word:
// 		logging.Debug("pass req to dlg", "st", "ok")
// 	default:
// 		logging.Warn("pass req to dlg", "st", "fail", "rsn", "ch closed or full")
// 		//TODO write good error
// 		return model.DictionaryEntry{}, fmt.Errorf("")
// 	}

// 	select {
// 	case resp := <-dialog.Resps():
// 		logging.Debug("get resp from dlg", "st", "ok")
// 		return resp, nil

// 	case <-time.After(10 * time.Second):
// 		logging.Warn("get resp from dlg", "st", "fail", "rsn", "timeout")

// 		//TODO write good error
// 		return model.DictionaryEntry{}, fmt.Errorf("task is cancelled")

// 	case <-ctx.Done():
// 		logging.Info("listen resp from dlg", "st", "finish", "rsn", "ctx done")

// 		//TODO write good error
// 		return model.DictionaryEntry{}, fmt.Errorf("task is cancelled")
// 	}
// }

func (s service) createButton(btn model.Button, id pkgmodel.ReqData) (button.Button, error) {
	if btn.Tp == model.AddToDictButton {
		out := addtodict.New(btn, s.dbs)

		out.UserId = uint64(id.DlgId)
		return out, nil
	}

	if btn.Tp == model.ExtraExamplesButton {
		out := extraexamples.New(btn)

		out.UserId = uint64(id.DlgId)
		return out, nil
	}

	if btn.Tp == model.MayBeMeantButton {
		out := maybemeant.New(btn, s.scn)

		out.UserId = uint64(id.DlgId)
		return out, nil
	}

	if btn.Tp == model.YouHeardRight {
		out := youheardright.New(btn, s.scn)

		out.UserId = uint64(id.DlgId)
		return out, nil
	}

	return nil, fmt.Errorf("not implemented")
}

func (s service) CreateButton(ctx context.Context, rData pkgmodel.ReqData, btn model.Button) (
	model.ButtonInfo, error) {

	button, err := s.createButton(btn, rData)
	if err != nil {
		//TODO asdas
		return model.ButtonInfo{}, err
	}

	id, err := s.repo.CreateInfo(btn)
	if err != nil {
		return model.ButtonInfo{}, err
	}

	//TODO fix this hook
	info := button.Info()
	info.Id = id

	return info, nil
}

func (s service) GetButton(ctx context.Context, rData pkgmodel.ReqData, id model.ButtonId) (
	model.ButtonInfo, error) {

	btn, err := s.repo.Info(id)
	if err != nil {
		return model.ButtonInfo{
			Id:   id,
			Text: "[Outdated]",
		}, nil
	}

	button, err := s.createButton(btn, rData)
	if err != nil {
		//TODO asdas
		return model.ButtonInfo{}, err
	}

	//TODO fix this hook
	info := button.Info()
	info.Id = id

	return info, nil

}

func (s service) ClickButton(ctx context.Context, rData pkgmodel.ReqData, id model.ButtonId) (
	model.ButtonInfo, error) {

	btn, err := s.repo.Info(id)
	if err != nil {
		return model.ButtonInfo{
			Id:   id,
			Text: "[Outdated]",
		}, nil
	}

	button, err := s.createButton(btn, rData)
	if err != nil {
		//TODO asdas
		return model.ButtonInfo{}, err
	}

	button.Click()

	//TODO fix this hook
	info := button.Info()
	info.Id = id

	return info, nil
}
