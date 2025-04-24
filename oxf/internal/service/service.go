package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"aidicti.top/oxf/internal/gateway/oxf"
	"aidicti.top/oxf/internal/model"
	"aidicti.top/pkg/gatewaydialog"
	pkgdialog "aidicti.top/pkg/gatewaydialog/dialog"
	pkgstorage "aidicti.top/pkg/gatewaydialog/storage"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
)

type gateway = gatewaydialog.Gateway[oxf.Req, oxf.Resp]
type storage = gatewaydialog.Storage[model.Word, model.DictionaryEntry]

type service struct {
	gtw gateway
	stg storage
}

func New(g gateway) *service {
	return &service{gtw: g}
}

type reqConverter struct {
	id pkgmodel.DlgID
}

func (c reqConverter) Convert(req model.Word) (oxf.Req, error) {
	reqId := pkgmodel.ReqID(rand.Uint64())

	return oxf.Req{Word: req, ReqData: pkgmodel.ReqData{DlgId: c.id, ReqId: reqId}}, nil
}

func (s *service) Run(ctx context.Context) {
	createFunc := func(id pkgmodel.DlgID) (
		gatewaydialog.Dialog[model.Word, model.DictionaryEntry], error) {

		cnv := reqConverter{id: id}

		logging.Debug("create dlg", "st", "ok", "id", id)

		return pkgdialog.New[model.DictionaryEntry](id, s.gtw, cnv), nil
	}

	s.stg = pkgstorage.New(ctx, createFunc)

	for {
		select {
		case resp := <-s.gtw.Resps():
			logging.Debug("get resp from gtw", "st", "ok", "dlgId", resp.DlgId, "reqId", resp.ReqId)

			dialog, err := s.stg.Get(resp.DlgId)
			if err != nil {
				logging.Warn("get dlg from stg", "st", "fail", "dlgId", resp.DlgId, "reqId", resp.ReqId, "err", err)
				break
			}

			select {
			case dialog.Resps() <- resp.DictionaryEntry:
				logging.Debug("pass resp to dlg", "st", "ok", "dlgId", resp.DlgId)

			default:
				logging.Warn("pass resp to dlg", "st", "fail", "rsn", "ch closed or full")
			}

		case <-ctx.Done():
			logging.Info("listen resp from gtw", "st", "finished", "rsn", "ctx done")
			return
		}
	}
}

func (s service) GetDictEntry(
	ctx context.Context,
	ids pkgmodel.ReqData,
	word model.Word) (model.DictionaryEntry, error) {

	dialog, err := s.stg.Get(ids.DlgId)
	if err != nil {
		logging.Warn("get dlg from stg", "st", "fail", "dlgId", ids.DlgId, "reqId", ids.ReqId, "err", err)

		return model.DictionaryEntry{}, err
	}

	select {
	case dialog.Reqs() <- word:
		logging.Debug("pass req to dlg", "st", "ok")
	default:
		logging.Warn("pass req to dlg", "st", "fail", "rsn", "ch closed or full")
		//TODO write good error
		return model.DictionaryEntry{}, fmt.Errorf("")
	}

	select {
	case resp := <-dialog.Resps():
		logging.Debug("get resp from dlg", "st", "ok")
		return resp, nil

	case <-time.After(10 * time.Second):
		logging.Warn("get resp from dlg", "st", "fail", "rsn", "timeout")

		//TODO write good error
		return model.DictionaryEntry{}, fmt.Errorf("task is cancelled")

	case <-ctx.Done():
		logging.Info("listen resp from dlg", "st", "finish", "rsn", "ctx done")

		//TODO write good error
		return model.DictionaryEntry{}, fmt.Errorf("task is cancelled")
	}
}
