package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"

	"aidicti.top/pkg/gatewaydialog"
	pkgdialog "aidicti.top/pkg/gatewaydialog/dialog"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/gatewaydialog/storage"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
	"aidicti.top/scn/internal/gateway/ai"
	"aidicti.top/scn/internal/gateway/dbs"
	"aidicti.top/scn/internal/gateway/oxf"
	"aidicti.top/scn/internal/gateway/sttg"
	"aidicti.top/scn/internal/gateway/uis"
	"aidicti.top/scn/internal/model"
	"github.com/google/uuid"

	model_oxf "aidicti.top/pkg/model/oxf"

	"aidicti.top/scn/internal/service/manager"
)

type STTGateway interface {
	Reqs() chan<- sttg.ReqSTT
	Resps() <-chan sttg.RespSTT
}

type TGGateway interface {
	Reqs() chan<- model.Message
	Resps() <-chan model.Message
}

type AIGateway interface {
	Reqs() chan<- ai.ReqAI
	Resps() <-chan ai.RespAI
}

type gatewayOXF interface {
	Reqs() chan<- oxf.Req
	Resps() <-chan oxf.Resp
}

type gatewayUIS interface {
	Reqs() chan<- uis.Req
	Resps() <-chan uis.Resp
}

type gatewayDBS interface {
	Reqs() chan<- dbs.Req
	Resps() <-chan dbs.Resp
}

type storageOXF = gatewaydialog.Storage[model_oxf.Word, model_oxf.DictionaryEntry]
type storageUIS = gatewaydialog.Storage[model.Button, model.ButtonId]
type storageDBS = gatewaydialog.Storage[dbs.Req, dbs.Resp]

type service struct {
	sttGtw STTGateway
	aiGtw  AIGateway
	tgGtw  TGGateway

	gtwOXF  gatewayOXF
	strgOXF storageOXF

	gtwUIS  gatewayUIS
	strgUIS storageUIS

	gtwDBS  gatewayDBS
	strgDBS storageDBS

	sessions map[model.UserID]Session
}

type Session interface {
	Run(context.Context)
}

type reqConverter struct {
	id pkgmodel.DlgID
}

func (c reqConverter) Convert(req model_oxf.Word) (oxf.Req, error) {
	reqId := pkgmodel.ReqID(rand.Uint64())

	return oxf.Req{Word: req, ReqData: pkgmodel.ReqData{DlgId: c.id, ReqId: reqId}}, nil
}

type reqConverter_uis struct {
	id pkgmodel.DlgID
}

func (c reqConverter_uis) Convert(btn model.Button) (uis.Req, error) {
	reqId := pkgmodel.ReqID(rand.Uint64())

	return uis.Req{
		ReqData: pkgmodel.ReqData{
			DlgId: c.id,
			ReqId: reqId},
		Button: btn,
	}, nil
}

type reqConverter_dbs struct {
	id pkgmodel.DlgID
}

func (c reqConverter_dbs) Convert(req dbs.Req) (dbs.Req, error) {
	reqId := pkgmodel.ReqID(rand.Uint64())

	req.ReqData = pkgmodel.ReqData{
		DlgId: c.id,
		ReqId: reqId}

	return req, nil
}

func New(stt STTGateway, ai AIGateway, tlg TGGateway, oxf gatewayOXF, uis gatewayUIS, dbs_ gatewayDBS) *service {
	return &service{
		sttGtw: stt,
		aiGtw:  ai,
		tgGtw:  tlg,
		gtwOXF: oxf,
		gtwUIS: uis,
		gtwDBS: dbs_,
		//TODO remove ctx from constructor and run as usual object
		strgOXF: storage.New(context.TODO(),
			func(id pkgmodel.DlgID) (
				gatewaydialog.Dialog[model_oxf.Word, model_oxf.DictionaryEntry], error) {

				cnv := reqConverter{id: id}

				logging.Debug("create dlg", "st", "ok", "id", id)

				return pkgdialog.New[model_oxf.DictionaryEntry](id, oxf, cnv), nil
			}),

		strgUIS: storage.New(context.TODO(),
			func(id pkgmodel.DlgID) (
				gatewaydialog.Dialog[model.Button, model.ButtonId], error) {

				cnv_ := reqConverter_uis{id: id}

				logging.Debug("create dlg", "st", "ok", "id", id)

				return pkgdialog.New[model.ButtonId](id, uis, cnv_), nil
			}),

		strgDBS: storage.New(context.TODO(),
			func(id pkgmodel.DlgID) (
				gatewaydialog.Dialog[dbs.Req, dbs.Resp], error) {

				cnv_ := reqConverter_dbs{id: id}

				logging.Debug("create dlg", "st", "ok", "id", id)

				return pkgdialog.New[dbs.Resp](id, dbs_, cnv_), nil
			}),

		sessions: make(map[model.UserID]Session),
	}
}

// type ReqConverter struct{}

// func (c ReqConverter) Convert(req model.AudioData) (sttg.ReqSTT, error) {
// 	return sttg.ReqSTT{}, nil
// }

type dialog[Resp any] interface {
	Resps() chan Resp
	Run(context.Context)
}

type GtwRespProcessor struct {
	mp map[model.DialogID]dialog[string]
	mx sync.Mutex
}

type UsrRespProcessor struct {
	mp map[model.DialogID]dialog[model.Message]
	mx sync.Mutex
	//TODO why its here?
	sess map[model.UserID]Session
	gtw  TGGateway
	//TODO nuff said
	sttRespProcessor *STTRespProcessor
	aiRespProcessor  *AIRespProcessor
	strgOXF          storageOXF
	strgUIS          storageUIS
	strgDBS          storageDBS
}

type STTRespProcessor struct {
	mp  map[model.DialogID]dialog[string]
	mx  sync.Mutex
	gtw STTGateway
}

func (c *STTRespProcessor) Process(ctx context.Context, resp sttg.RespSTT) error {
	getDialog := func() (dialog[string], error) {
		c.mx.Lock()
		defer c.mx.Unlock()

		d, exists := c.mp[model.DialogID(resp.DialogID)]

		if !exists {
			return nil, fmt.Errorf("dialog not exist")
		}

		return d, nil
	}

	d, err := getDialog()

	if err != nil {
		logging.Info("dialog not exist", "id", resp.DialogID)

		return err
	}

	d.Resps() <- resp.Data

	logging.Debug("gateway pass resp")

	return nil
}

type AIRespProcessor struct {
	mp  map[model.DialogID]dialog[ai.RespAI]
	mx  sync.Mutex
	gtw AIGateway
}

func (c *AIRespProcessor) Process(ctx context.Context, resp ai.RespAI) error {
	//TODO finding dialog its common part to generic
	getDialog := func() (dialog[ai.RespAI], error) {
		c.mx.Lock()
		defer c.mx.Unlock()

		d, exists := c.mp[model.DialogID(resp.Id.DlgId)]

		if !exists {
			return nil, fmt.Errorf("dialog not exist")
		}

		return d, nil
	}

	d, err := getDialog()

	if err != nil {
		logging.Info("dialog not exist", "id", resp.Id.DlgId)

		return err
	}

	d.Resps() <- resp

	logging.Debug("gateway pass resp")

	return nil
}

type OXFRespProcessor struct {
	strg storageOXF
}

func (c *OXFRespProcessor) Process(ctx context.Context, resp oxf.Resp) error {
	dialog, err := c.strg.Get(resp.DlgId)
	if err != nil {
		logging.Warn("get dlg from stg", "st", "fail", "dlgId", resp.DlgId, "reqId", resp.ReqId, "err", err)
		return err
	}

	select {
	case dialog.Resps() <- resp.DictionaryEntry:
		logging.Debug("pass resp to dlg", "st", "ok", "dlgId", resp.DlgId)

	default:
		logging.Warn("pass resp to dlg", "st", "fail", "rsn", "ch closed or full")
		return fmt.Errorf("asdasdasdasdasd")
	}

	return nil
}

type UISRespProcessor struct {
	strg storageUIS
}

func (c *UISRespProcessor) Process(ctx context.Context, resp uis.Resp) error {
	dialog, err := c.strg.Get(resp.DlgId)
	if err != nil {
		logging.Warn("get dlg from stg", "st", "fail", "dlgId", resp.DlgId, "reqId", resp.ReqId, "err", err)
		return err
	}

	select {
	case dialog.Resps() <- resp.ButtonId:
		logging.Debug("pass resp to dlg", "st", "ok", "dlgId", resp.DlgId)

	default:
		logging.Warn("pass resp to dlg", "st", "fail", "rsn", "ch closed or full")
		return fmt.Errorf("asdasdasdasdasd")
	}

	return nil
}

type DBSRespProcessor struct {
	strg storageDBS
}

func (c *DBSRespProcessor) Process(ctx context.Context, resp dbs.Resp) error {
	dialog, err := c.strg.Get(resp.DlgId)
	if err != nil {
		logging.Warn("get dlg from stg", "st", "fail", "dlgId", resp.DlgId, "reqId", resp.ReqId, "err", err)
		return err
	}

	select {
	case dialog.Resps() <- resp:
		logging.Debug("pass resp to dlg", "st", "ok", "dlgId", resp.DlgId)

	default:
		logging.Warn("pass resp to dlg", "st", "fail", "rsn", "ch closed or full")
		return fmt.Errorf("asdasdasdasdasd")
	}

	return nil
}

type DialogReqProcessor struct{}

func (c *DialogReqProcessor) Process(ctx context.Context, resp model.Message) {

}

type TGReqConverter struct {
	id model.UserID
}

func (c TGReqConverter) Convert(req model.Message) (model.Message, error) {
	//TODO we should modify message model.Message to put real id, thats a hook,
	//we need to create model.Message analog without id for dialog, and put user id here

	req.Id = c.id
	return req, nil
}

type STTReqConverter struct {
	id model.DialogID
}

func (c STTReqConverter) Convert(req model.AudioData) (sttg.ReqSTT, error) {
	utils.Assert(len(req.Data) != 0, "req size == 0")

	return sttg.ReqSTT{
		ReqId:     model.RequestID(c.id),
		DialogID:  c.id,
		AudioData: req,
	}, nil
}

type AIReqConverter struct {
	id model.DialogID
}

func (c AIReqConverter) Convert(req ai.ReqAI) (ai.ReqAI, error) {
	// utils.Assert(len(req.Words) != 0, "req size == 0")

	reqId := utils.Must(uuid.NewRandom())

	return ai.ReqAI{
		Id: pkgmodel.ReqData{
			ReqId: pkgmodel.ReqID(reqId.ID()),
			DlgId: pkgmodel.DlgID(c.id),
		},
		Words:            req.Words,
		SentencesToCheck: req.SentencesToCheck,
	}, nil
}

//TODO we may resolve a problem with creating just making a new processor that creates a new session and passing message furter

func (c *UsrRespProcessor) Process(ctx context.Context, resp model.Message) error {
	getDialog := func() (dialog[model.Message], error) {
		c.mx.Lock()
		defer c.mx.Unlock()

		d, exists := c.mp[model.DialogID(resp.Id)]

		if !exists {
			return nil, fmt.Errorf("dialog not exist")
		}

		return d, nil
	}

	d, err := getDialog()

	if err != nil {
		logging.Info("dialog not exist", "id", resp.Id)

		if _, ok := c.sess[resp.Id]; !ok {
			d := pkgdialog.New[model.Message](pkgmodel.DlgID(resp.Id), c.gtw, TGReqConverter{resp.Id})
			go d.Run(ctx)

			dstt := pkgdialog.New[string](pkgmodel.DlgID(resp.Id), c.sttRespProcessor.gtw, STTReqConverter{model.DialogID(resp.Id)})
			go dstt.Run(ctx)

			dai := pkgdialog.New[ai.RespAI](pkgmodel.DlgID(resp.Id), c.aiRespProcessor.gtw, AIReqConverter{model.DialogID(resp.Id)})
			go dai.Run(ctx)

			doxf, err := c.strgOXF.Get(pkgmodel.DlgID(resp.Id))
			//TODO i sure there is may be an error also we should change assert interface to take any...
			utils.Assert(err == nil, "")

			duis, err := c.strgUIS.Get(pkgmodel.DlgID(resp.Id))
			utils.Assert(err == nil, "")

			ddbs, err := c.strgDBS.Get(pkgmodel.DlgID(resp.Id))
			utils.Assert(err == nil, "")

			c.mx.Lock()
			c.aiRespProcessor.mx.Lock()
			c.sttRespProcessor.mx.Lock()

			c.mp[model.DialogID(resp.Id)] = d
			c.aiRespProcessor.mp[model.DialogID(resp.Id)] = dai
			c.sttRespProcessor.mp[model.DialogID(resp.Id)] = dstt

			c.sess[resp.Id] = manager.New(d, dai, dstt, doxf, duis, ddbs)
			go c.sess[resp.Id].Run(ctx)

			c.sttRespProcessor.mx.Unlock()
			c.aiRespProcessor.mx.Unlock()
			c.mx.Unlock()
		}

		d, _ = getDialog()
	}

	select {
	case d.Resps() <- resp:
		logging.Debug("passed resp to dialog")
	default:
		logging.Debug("dialog not listen resp")
	}

	logging.Debug("gateway pass resp")

	return nil
}

type nullConv struct{}

func (c nullConv) Convert(struct{}) (struct{}, error) {
	return struct{}{}, nil
}

func (s service) Run(ctx context.Context) {
	strgStt := STTRespProcessor{
		mp:  make(map[model.DialogID]dialog[string]),
		gtw: s.sttGtw,
	}

	lstnStt := listener.New(s.sttGtw.Resps(), &strgStt)

	go lstnStt.Run(ctx)

	strgAi := AIRespProcessor{
		mp:  make(map[model.DialogID]dialog[ai.RespAI]),
		gtw: s.aiGtw,
	}

	lstnAi := listener.New(s.aiGtw.Resps(), &strgAi)

	go lstnAi.Run(ctx)

	strgUsr := UsrRespProcessor{
		mp:               make(map[model.DialogID]dialog[model.Message]),
		sess:             make(map[model.UserID]Session),
		gtw:              s.tgGtw,
		sttRespProcessor: &strgStt,
		aiRespProcessor:  &strgAi,
		strgOXF:          s.strgOXF,
		strgUIS:          s.strgUIS,
		strgDBS:          s.strgDBS,
	}

	//TODO i need a new listener tg to create service if there is no any
	lstnUsr := listener.New(s.tgGtw.Resps(), &strgUsr)
	go lstnUsr.Run(ctx)

	procRespOXF := OXFRespProcessor{s.strgOXF}
	lstnOxf := listener.New(s.gtwOXF.Resps(), &procRespOXF)
	go lstnOxf.Run(ctx)

	procRespUIS := UISRespProcessor{s.strgUIS}
	lstnUIS := listener.New(s.gtwUIS.Resps(), &procRespUIS)
	go lstnUIS.Run(ctx)

	procRespDBS := DBSRespProcessor{s.strgDBS}
	lstnDBS := listener.New(s.gtwDBS.Resps(), &procRespDBS)
	go lstnDBS.Run(ctx)

	select {
	case <-ctx.Done():
	}
}
