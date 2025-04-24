package translatesentences

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/utils"
	"aidicti.top/scn/internal/gateway/ai"
	"aidicti.top/scn/internal/gateway/dbs"
	"aidicti.top/scn/internal/model"
	"aidicti.top/scn/internal/service/scenario"
	"aidicti.top/scn/internal/service/scenario/transcriptaudio"
)

type scnr struct {
	usr scenario.DialogUSR
	stt scenario.DialogSTT
	gpt scenario.DialogGPT
	oxf scenario.DialogOXF
	uis scenario.DialogUIS
	dbs scenario.DialogDBS

	name string
}

type dialog[Req any, Resp any] interface {
	Reqs() chan Req
	Resps() chan Resp
}

type intermediateTimers[Req any, Resp any] struct {
	reqs  chan Req
	resps chan Resp

	timeMsgs []timeMsg
	d        dialog[Req, Resp]
}

type timeMsg struct {
	Time time.Duration
	Fs   func(context.Context)
}

func NewIntermediateTimers[Req any, Resp any](dlg dialog[Req, Resp], timeMsgs []timeMsg) *intermediateTimers[Req, Resp] {

	utils.Contract(len(timeMsgs) > 0)

	return &intermediateTimers[Req, Resp]{
		reqs:     make(chan Req),
		resps:    make(chan Resp),
		timeMsgs: timeMsgs,
		d:        dlg,
	}
}

func (t intermediateTimers[Req, Resp]) Reqs() chan Req {
	return t.reqs
}

func (t intermediateTimers[Req, Resp]) Resps() chan Resp {
	return t.resps
}

func (t intermediateTimers[Req, Resp]) Run(ctx context.Context) {
	defer close(t.reqs)

	select {
	case req := <-t.reqs:
		logging.Debug("get req for inter", "st", "ok")

		select {
		case t.d.Reqs() <- req:
			logging.Debug("get req for inter", "st", "ok")

		case <-ctx.Done():
			logging.Debug("get req for inter", "st", "finish", "rsn", "ctx done")
			return
		}

	case <-ctx.Done():
		logging.Debug("get req for inter", "st", "finish", "rsn", "ctx done")
		return
	}

	timers := t.timeMsgs

	for {
		select {
		case resp := <-t.d.Resps():
			logging.Debug("get req for inter", "st", "finish", "rsn", "ctx done")

			select {
			case t.resps <- resp:
				logging.Debug("get req for inter", "st", "ok")
				return

			case <-ctx.Done():
				logging.Debug("get req for inter", "st", "finish", "rsn", "ctx done")
				return
			}

		case <-time.After(timers[0].Time):
			// utils.Contract(false)

			logging.Debug("TODO timer", "st", "finish", "rsn", "ctx done")
			timers[0].Fs(ctx)

			logging.Debug("TODO done", "st", "finish", "rsn", "ctx done")

			if len(timers) == 1 {
				return
			}

			timers = timers[1:]

		case <-ctx.Done():
			logging.Debug("TODO done", "st", "finish", "rsn", "ctx done")
			return
		}
	}

}

func New(
	stt scenario.DialogSTT,
	usr scenario.DialogUSR,
	gpt scenario.DialogGPT,
	oxf scenario.DialogOXF,
	uis scenario.DialogUIS,
	dbs scenario.DialogDBS) *scnr {

	return &scnr{usr: usr, stt: stt, gpt: gpt, oxf: oxf, uis: uis, dbs: dbs, name: "translatesentences"}
}

func (s scnr) Name() string {
	return s.name
}

func (s scnr) Run(ctx context.Context) {
	select {
	case s.dbs.Reqs() <- dbs.Req{
		DictWord: dbs.DictWord{},
		Method:   dbs.Get}:

		logging.Debug("send req to dbs", "st", "ok", "type", s.name)

	case <-time.After(1 * time.Second):
		logging.Debug("send req to dbs", "st", "fail", "rsn", "timeout", "type", s.name)
		return

	case <-ctx.Done():
		logging.Debug("send req to dbs", "st", "ok", "type", s.name)
		return
	}

	var dbsResp = dbs.Resp{}
	select {
	case dbsResp = <-s.dbs.Resps():
		logging.Debug("get resp from dbs", "st", "ok", "type", s.name)

	case <-time.After(2 * time.Second):
		logging.Debug("get resp from dbs", "st", "fail", "rsn", "timeout", "type", s.name)
		return

	case <-ctx.Done():
		logging.Debug("get resp from dbs", "st", "fail", "rsn", "ctx done", "type", s.name)
		return
	}

	if len(dbsResp.Words) < 1 {
		logging.Debug("TODO not enoght to get request", "type", s.name)

		s.usr.Reqs() <- model.Message{
			Message: "oof, please, add more words, there is no enough!!!",
		}

		return
	}

	pickWords := func(words []dbs.DictWord) []model.WordReq {
		sum := uint64(0)
		for _, word := range words {
			sum += uint64(word.ID)
		}

		out := []model.WordReq{}

		// countWords := 3 + (rand.Uint() % 3)
		countWords := 1

		for _ = range countWords {
			rnd := uint64(rand.Uint64() % sum)

			localSum := uint64(0)

			for _, word := range words {
				localSum += uint64(word.ID)
				if localSum > rnd {
					out = append(out, model.WordReq{
						Word: word.Word,
						Info: word.Def,
					})
					break
				}
			}
		}

		return out
	}

	igpt := NewIntermediateTimers(s.gpt, []timeMsg{
		{3 * time.Second, func(ctx context.Context) {
			s.usr.Reqs() <- model.Message{
				Message: "Still here, just waiting for ChatGPT to finish its coffee break.",
			}
		},
		},
		{3 * time.Second, func(ctx context.Context) {
			s.usr.Resps() <- model.Message{
				Message: "ChatGPT is thinking really hard... or just staring into the void.",
			}
		},
		},
		{3 * time.Second, func(ctx context.Context) {
			s.usr.Resps() <- model.Message{
				Message: "The AI overlord is busy—probably negotiating with Skynet.",
			}
		},
		},
		{3 * time.Second, func(ctx context.Context) {
			s.usr.Resps() <- model.Message{
				Message: "Brb, teaching ChatGPT the concept of 'hurry up'.",
			}
		},
		},
	})

	go igpt.Run(ctx)

	select {
	case igpt.Reqs() <- ai.ReqAI{
		// case s.gpt.Reqs() <- ai.ReqAI{
		Words: &model.WordsReq{
			Words: pickWords(dbsResp.Words),
		}}:

		logging.Debug("send req to gpt", "st", "ok", "type", s.name)

	case <-time.After(3 * time.Second):
		logging.Debug("send req to gpt", "st", "fail", "rsn", "timeout", "type", s.name)
		return

	case <-ctx.Done():
		logging.Debug("get resp from gpt", "st", "fail", "rsn", "ctx done", "type", s.name)
		return
	}

	var gptResp = ai.RespAI{}
	select {
	case gptResp = <-igpt.Resps():
		// case gptResp = <-s.gpt.Resps():
		logging.Debug("get resp from gpt", "st", "ok", "type", s.name)

	case <-time.After(15 * time.Second):
		logging.Debug("get resp from gpt", "st", "fail", "rsn", "timeout", "type", s.name)
		return

	case <-ctx.Done():
		logging.Debug("get resp from gpt", "st", "fail", "rsn", "ctx done", "type", s.name)
		return
	}

	s.usr.Reqs() <- model.Message{
		Message: "ok, try transalte these sentences, may use voice message!",
	}

	//TODO gptResp.Sents.Sents may panic
	for id, sent := range gptResp.Sents.Sents {
		s.usr.Reqs() <- model.Message{
			Message: fmt.Sprintf("%d. %s <span class=\"tg-spoiler\">%s</span>", id, sent.Original, sent.Word),
		}
	}

	answer := ""

	for answer == "" {

		sntsAns := model.Message{}
		select {
		case sntsAns = <-s.usr.Resps():
			logging.Debug("get resp from usr", "st", "ok", "type", s.name)

		case <-ctx.Done():
			logging.Debug("get resp from usr", "st", "fail", "rsn", "ctx done", "type", s.name)
			return
		}

		if sntsAns.Audio != nil {
			s.usr.Reqs() <- model.Message{
				Message: "ok, give me some time!",
			}

			cmd := transcriptaudio.New(s.stt, sntsAns.Audio.Data)

			cmd.Run(ctx)

			jsonData, err := json.Marshal(struct {
				Txt string `json:"txt"`
			}{
				Txt: cmd.Result,
			})
			if err != nil {
				logging.Debug("marshal data ", "st", "fail", "type", s.name)
				return
			}

			select {
			case s.uis.Reqs() <- model.Button{
				Tp:    model.YouHeardRight,
				Texts: []string{"you're right!"},
				Meta:  jsonData,
			}:
				logging.Debug("send req to uis", "st", "ok", "type", s.name)

				select {
				case id_btn := <-s.uis.Resps():
					logging.Debug("get resp from uis", "st", "ok", "type", s.name)

					select {
					case s.usr.Reqs() <- model.Message{
						Message: fmt.Sprintf("that what i heard: \n%s", cmd.Result),
						Actions: []uint64{uint64(id_btn)}}:

						logging.Debug("TODO", "st", "ok", "type", s.name)

					case <-ctx.Done():
						logging.Debug("TODO", "st", "ok", "type", s.name)

					}

				case <-time.After(1 * time.Second):
					logging.Debug("get resp from uis", "st", "fail", "rsn", "timeout", "type", s.name)

				case <-ctx.Done():
					logging.Debug("get resp from uis", "st", "ok", "type", s.name)
					return
				}

			case <-time.After(1 * time.Second):
				logging.Debug("send req to uis", "st", "fail", "rsn", "timeout", "type", s.name)

			case <-ctx.Done():
				logging.Debug("send req to uis", "st", "fail", "rsn", "ctx done", "type", s.name)
				return
			}

		} else {
			answer = sntsAns.Message
		}

	}

	s.usr.Reqs() <- model.Message{Message: answer}

	select {
	case s.gpt.Reqs() <- ai.ReqAI{SentencesToCheck: answer}:
		logging.Debug("send req to gpt", "st", "ok", "type", s.name)

	case <-time.After(15 * time.Second):
		logging.Debug("send req to gpt", "st", "fail", "rsn", "timeout", "type", s.name)
		return

	case <-ctx.Done():
		logging.Debug("get resp from gpt", "st", "fail", "rsn", "ctx done", "type", s.name)
		return
	}

	select {
	case gptResp = <-s.gpt.Resps():
		logging.Debug("get resp from gpt", "st", "ok", "type", s.name)

	case <-time.After(10 * time.Second):
		logging.Debug("get resp from gpt", "st", "fail", "rsn", "timeout", "type", s.name)
		return

	case <-ctx.Done():
		logging.Debug("get resp from gpt", "st", "fail", "rsn", "ctx done", "type", s.name)
		return
	}

	if gptResp.Results == nil {
		s.usr.Reqs() <- model.Message{
			Message: "oops, ai answers something, but it's something wrong",
		}

		return
	}

	for _, result := range gptResp.Results.Results {
		s.usr.Reqs() <- model.Message{
			Message: fmt.Sprintf("Rating - %d\n %s\n Best: %s", result.Rating, result.Explaination, result.Correction),
		}
	}
}
