package session

import (
	"context"
	"sync"

	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/model/oxf"
	"aidicti.top/pkg/utils"
	"aidicti.top/scn/internal/gateway/ai"
	"aidicti.top/scn/internal/gateway/dbs"
	"aidicti.top/scn/internal/model"
	"aidicti.top/scn/internal/service/scenario"
	"github.com/google/uuid"
)

type dialogSTT interface {
	Reqs() chan model.AudioData
	Resps() chan string
}

type sttGateway interface {
	NewDialog(id model.DialogID) dialogSTT
}

type dialogUSR interface {
	Reqs() chan model.Message
	Resps() chan model.Message
}

type usrGateway interface {
	NewDialog(id model.DialogID) dialogUSR
}

type DialogAI interface {
	Reqs() chan ai.ReqAI
	Resps() chan ai.RespAI
}

type aiGateway interface {
	// NewDialog(id model.DialogID) sttDialog
}

type dialog[Req any, Resp any] struct {
	id    model.DialogID
	reqs  chan Req
	resps chan Resp
}

func (d dialog[Req, Resp]) Reqs() chan Req {
	return d.reqs
}

func (d dialog[Req, Resp]) Resps() chan Resp {
	return d.resps
}

func newDialog[Req any, Resp any](id model.DialogID, reqs chan Req, resps chan Resp) *dialog[Req, Resp] {
	return &dialog[Req, Resp]{id: id, reqs: reqs, resps: resps}
}

type session struct {
	id model.SessionID

	stt, scnrSTT dialogSTT
	usr, scnrUSR dialogUSR
	gpt, scnrGPT DialogAI
	oxf, scnrOXF scenario.DialogOXF
	uis, scnrUIS scenario.DialogUIS
	dbs, scnrDBS scenario.DialogDBS

	scnr scenario.Scenario

	mtx sync.Mutex
}

func (s *session) RunScenario(
	ctx context.Context,
	create func(
		stt scenario.DialogSTT,
		usr scenario.DialogUSR,
		gpt scenario.DialogGPT,
		oxf scenario.DialogOXF,
		uis scenario.DialogUIS,
		dbs scenario.DialogDBS) scenario.Scenario) {

	logging.Debug("set scenario to session", "id", s.id)

	s.mtx.Lock()

	if (s.scnrSTT != nil) && (s.scnrUIS != nil) &&
		(s.scnrUSR != nil) && (s.scnrGPT != nil) && (s.scnrOXF != nil) && (s.scnrDBS != nil) {

		close(s.scnrSTT.Resps())
		close(s.scnrUSR.Resps())
		close(s.scnrGPT.Resps())
		close(s.scnrOXF.Resps())
		close(s.scnrUIS.Resps())
		close(s.scnrDBS.Resps())
	}

	uuidId, err := uuid.NewRandom()
	utils.Assert(err == nil, "cant create uuid")

	id := model.DialogID(uuidId.ID())

	s.scnrSTT = newDialog(id, make(chan model.AudioData), make(chan string))
	s.scnrUSR = newDialog(id, make(chan model.Message), make(chan model.Message))
	s.scnrGPT = newDialog(id, make(chan ai.ReqAI), make(chan ai.RespAI))
	s.scnrOXF = newDialog(id, make(chan oxf.Word), make(chan oxf.DictionaryEntry))
	s.scnrUIS = newDialog(id, make(chan model.Button), make(chan model.ButtonId))
	s.scnrDBS = newDialog(id, make(chan dbs.Req), make(chan dbs.Resp))

	s.scnr = create(s.scnrSTT, s.scnrUSR, s.scnrGPT, s.scnrOXF, s.scnrUIS, s.scnrDBS)

	go func() {
		for {
			select {
			case <-ctx.Done():
				logging.Info("listen to dlg", "st", "finish", "rsn", "ctx done", "id", id)
				return

			case req := <-s.scnrUSR.Reqs():
				logging.Debug("get usr req from dlg", "st", "ok", "id", s.id)

				select {
				case s.usr.Reqs() <- req:
					logging.Debug("pass usr req from dlg", "st", "ok", "id", s.id)
				default:
					logging.Info("pass usr req from dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
				}

			case req := <-s.scnrGPT.Reqs():
				logging.Debug("get gpt req", "st", "ok", "id", s.id)

				select {
				case s.gpt.Reqs() <- req:
					logging.Debug("pass gpt req from dlg", "st", "ok", "id", s.id)
				default:
					logging.Info("pass gpt req from dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
				}

			case req := <-s.scnrSTT.Reqs():
				logging.Debug("get stt resp", "st", "ok", "id", s.id)

				select {
				case s.stt.Reqs() <- req:
					logging.Debug("pass stt req from dlg", "st", "ok", "id", s.id)
				default:
					logging.Info("pass stt req from dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
				}

			case req := <-s.scnrOXF.Reqs():
				logging.Debug("get oxf resp", "st", "ok", "id", s.id)

				select {
				case s.oxf.Reqs() <- req:
					logging.Debug("pass oxf req from dlg", "st", "ok", "id", s.id)
				default:
					logging.Info("pass oxf req from dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
				}

			case req := <-s.scnrUIS.Reqs():
				logging.Debug("get uis resp", "st", "ok", "id", s.id)

				select {
				case s.uis.Reqs() <- req:
					logging.Debug("pass uis req from dlg", "st", "ok", "id", s.id)
				default:
					logging.Info("pass uis req from dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
				}

			case req := <-s.scnrDBS.Reqs():
				logging.Debug("get dbs resp", "st", "ok", "id", s.id)

				select {
				case s.dbs.Reqs() <- req:
					logging.Debug("pass dbs req from dlg", "st", "ok", "id", s.id)
				default:
					logging.Info("pass dbs req from dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
				}
			}
		}
	}()

	s.mtx.Unlock()

	s.scnr.Run(ctx)
}

// TODO need for scenario manager
func New(
	stt dialogSTT,
	usr dialogUSR,
	ai DialogAI,
	oxf scenario.DialogOXF,
	uis scenario.DialogUIS,
	dbs scenario.DialogDBS) *session {

	uuidId, err := uuid.NewRandom()
	utils.Assert(err == nil, "cant create uuid")

	id := model.SessionID(uuidId.ID())

	s := &session{id: id, stt: stt, usr: usr, gpt: ai, scnr: nil, oxf: oxf, uis: uis, dbs: dbs}

	logging.Debug("session created", "id", id)
	return s
}

func (s *session) Run(ctx context.Context) {
	utils.Assert(s.scnr != nil, "")

	for {
		logging.Debug("run session", "st", "start", "id", s.id)

		select {
		case <-ctx.Done():
			logging.Debug("run session", "st", "finish", "rsn", "ctx done", "id", s.id)

			return

		case resp := <-s.usr.Resps():
			logging.Debug("get usr resp", "st", "ok", "id", s.id)

			s.mtx.Lock()

			select {
			case s.scnrUSR.Resps() <- resp:
				logging.Debug("pass usr resp to dlg", "st", "ok", "id", s.id)
			default:
				logging.Info("pass usr resp to dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
			}

			s.mtx.Unlock()

		case resp := <-s.gpt.Resps():
			logging.Debug("get gpt resp", "st", "ok", "id", s.id)

			s.mtx.Lock()

			select {
			case s.scnrGPT.Resps() <- resp:
				logging.Debug("pass gpt resp to dlg", "st", "ok", "id", s.id)
			default:
				logging.Info("pass gpt resp to dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
			}

			s.mtx.Unlock()

		case resp := <-s.stt.Resps():
			logging.Debug("get stt resp", "st", "ok", "id", s.id)

			s.mtx.Lock()

			select {
			case s.scnrSTT.Resps() <- resp:
				logging.Debug("pass stt resp to dlg", "st", "ok", "id", s.id)
			default:
				logging.Info("pass stt resp to dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
			}

			s.mtx.Unlock()

		case resp := <-s.oxf.Resps():
			logging.Debug("get oxf resp", "st", "ok", "id", s.id)

			s.mtx.Lock()

			select {
			case s.scnrOXF.Resps() <- resp:
				logging.Debug("pass oxf resp to dlg", "st", "ok", "id", s.id)
			default:
				logging.Info("pass oxf resp to dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
			}

			s.mtx.Unlock()

		case resp := <-s.uis.Resps():
			logging.Debug("get uis resp", "st", "ok", "id", s.id)

			s.mtx.Lock()

			select {
			case s.scnrUIS.Resps() <- resp:
				logging.Debug("pass uis resp to dlg", "st", "ok", "id", s.id)
			default:
				logging.Info("pass uis resp to dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
			}

			s.mtx.Unlock()

		case resp := <-s.dbs.Resps():
			logging.Debug("get dbs resp", "st", "ok", "id", s.id)

			s.mtx.Lock()

			select {
			case s.scnrDBS.Resps() <- resp:
				logging.Debug("pass dbs resp to dlg", "st", "ok", "id", s.id)
			default:
				logging.Info("pass dbs resp to dlg", "st", "fail", "rsn", "dlg not listen", "id", s.id)
			}

			s.mtx.Unlock()
		}
	}
}
