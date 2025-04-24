package manager

import (
	"context"
	"sync"
	"time"

	"aidicti.top/pkg/logging"
	"aidicti.top/scn/internal/model"
	"aidicti.top/scn/internal/service/scenario"
	"aidicti.top/scn/internal/service/scenario/getdictentry"
	"aidicti.top/scn/internal/service/scenario/idle"
	"aidicti.top/scn/internal/service/scenario/translatesentences"
	"aidicti.top/scn/internal/service/session"
)

type manager struct {
	usr scenario.DialogUSR
	gpt scenario.DialogGPT
	stt scenario.DialogSTT
	oxf scenario.DialogOXF
	uis scenario.DialogUIS
	dbs scenario.DialogDBS

	real_usr scenario.DialogUSR
}

type dialog struct {
	reqs  chan model.Message
	resps chan model.Message
}

func (d dialog) Reqs() chan model.Message {
	return d.reqs
}

func (d dialog) Resps() chan model.Message {
	return d.resps
}

func New(
	usr scenario.DialogUSR,
	gpt scenario.DialogGPT,
	stt scenario.DialogSTT,
	oxf scenario.DialogOXF,
	uis scenario.DialogUIS,
	dbs scenario.DialogDBS) *manager {

	return &manager{
		usr:      dialog{reqs: usr.Reqs(), resps: make(chan model.Message)},
		gpt:      gpt,
		stt:      stt,
		oxf:      oxf,
		real_usr: usr,
		uis:      uis,
		dbs:      dbs,
	}
}

func (m manager) Run(ctx context.Context) {
	s := session.New(m.stt, m.usr, m.gpt, m.oxf, m.uis, m.dbs)

	createIdle := func(
		scenario.DialogSTT,
		scenario.DialogUSR,
		scenario.DialogGPT,
		scenario.DialogOXF,
		scenario.DialogUIS,
		scenario.DialogDBS) scenario.Scenario {

		return idle.New()
	}

	_ = createIdle

	createGetDictEntry := func(
		stt scenario.DialogSTT,
		usr scenario.DialogUSR,
		gpt scenario.DialogGPT,
		oxf scenario.DialogOXF,
		uis scenario.DialogUIS,
		_ scenario.DialogDBS) scenario.Scenario {

		return getdictentry.New(stt, usr, gpt, oxf, uis)
	}

	createTranslateSent := func(
		stt scenario.DialogSTT,
		usr scenario.DialogUSR,
		gpt scenario.DialogGPT,
		oxf scenario.DialogOXF,
		uis scenario.DialogUIS,
		dbs scenario.DialogDBS) scenario.Scenario {

		return translatesentences.New(stt, usr, gpt, oxf, uis, dbs)
	}

	var cancel *context.CancelFunc
	mtx := sync.Mutex{}

	// barrier := make(chan struct{}, 0)
	// once := sync.Once{}

	scnChan := make(
		chan func(scenario.DialogSTT,
			scenario.DialogUSR,
			scenario.DialogGPT,
			scenario.DialogOXF,
			scenario.DialogUIS,
			scenario.DialogDBS) scenario.Scenario, 1)

	go func(ctx context.Context) {
		for {
			mtx.Lock()

			scnCtx, cancelCtx := context.WithCancel(ctx)

			cancel = &cancelCtx

			select {
			case <-ctx.Done():
				logging.Info("mng set getdictentry", "rsn", "command")

				mtx.Unlock()

				return

			case scn := <-scnChan:
				mtx.Unlock()

				s.RunScenario(scnCtx, scn)

				logging.Info("scn done", "rsn", "command")

			default:
				select {
				case scnChan <- createGetDictEntry:
					mtx.Unlock()

					logging.Info("mng set getdictentry", "rsn", "command")

				default:
					mtx.Unlock()

					logging.Info("mng set getdictentry", "rsn", "command")
				}
			}
		}

	}(ctx)

	scnChan <- createGetDictEntry

	//TODO hook to wait until scn is set
	<-time.After(200 * time.Millisecond)

	go s.Run(ctx)

	for {
		select {
		case msg := <-m.real_usr.Resps():
			if msg.Message == "/dict" {
				logging.Info("mng set getdictentry", "rsn", "command")

				mtx.Lock()

				(*cancel)()

				scnChan <- createGetDictEntry

				mtx.Unlock()

				logging.Debug("mng set getdictentry", "st", "finish", "rsn", "command")

				break
			}

			if msg.Message == "/pract" {
				logging.Info("mng set translatesentences", "rsn", "command")

				mtx.Lock()

				(*cancel)()

				scnChan <- createTranslateSent

				mtx.Unlock()

				logging.Debug("mng set translatesentences", "st", "finish", "rsn", "command")

				break
			}

			select {
			case m.usr.Resps() <- msg:
				logging.Debug("TODO passing msg to session")
			default:
				logging.Debug("TODO passing msg to session fail not listen")

				//TODO hook to wait until scn is set
				<-time.After(200 * time.Millisecond)

				select {
				case m.usr.Resps() <- msg:
					logging.Debug("TODO passing msg to session")
				default:
					logging.Debug("TODO passing msg to session fail not listen")
				}
			}
		}
	}
}
