package createsentences

import (
	"context"
	"fmt"

	"aidicti.top/pkg/logging"
	"aidicti.top/scn/internal/gateway/ai"
	"aidicti.top/scn/internal/model"
	"aidicti.top/scn/internal/service/scenario"
)

type createsentences struct {
	ai    scenario.DialogGPT
	usr   scenario.DialogUSR
	stt   scenario.DialogSTT
	words []string

	name string
}

func New(
	ai scenario.DialogGPT,
	usr scenario.DialogUSR,
	stt scenario.DialogSTT,
	words []string) *createsentences {

	return &createsentences{ai: ai, usr: usr, stt: stt, words: words, name: "createsentences"}
}

func (scnr createsentences) Run(ctx context.Context) {
	Words := model.WordsReq{}

	for _, value := range scnr.words {
		Words.Words = append(Words.Words, model.WordReq{
			Word: value,
			Info: "",
		})
	}

	select {
	case scnr.ai.Reqs() <- ai.ReqAI{Words: &Words}:
		logging.Debug("ai sent sent")
	case <-ctx.Done():
		logging.Debug("ai sent closed by context")
		return
	}

	sentencesToTranslate := ai.RespAI{}
	select {
	case answ := <-scnr.ai.Resps():
		sentencesToTranslate = answ
		logging.Debug("ai got resp", "respId", answ, "respMsg", answ)

	case <-ctx.Done():
		logging.Debug("ai got resp")
		return
	}

	//TODO use string builder
	userRequestToTranslate := ""
	for i, sents := range sentencesToTranslate.Sents.Sents {
		userRequestToTranslate = userRequestToTranslate + fmt.Sprintf("%d.", i) + " " + sents.Original + "\n"
	}

	select {
	case scnr.usr.Reqs() <- model.Message{Message: userRequestToTranslate}:
		logging.Debug("usr sent sent")
	case <-ctx.Done():
		logging.Debug("usr sent closed by context")
		return
	}

	translated := model.Message{}
	select {
	case translated = <-scnr.usr.Resps():
		logging.Debug("get resp from usr", "st", "ok", "type", scnr.name)

	case <-ctx.Done():
		logging.Debug("get resp from usr", "st", "fail", "rsn", "ctx done")
		return
	}

	if translated.Audio != nil {
		select {
		case scnr.usr.Reqs() <- model.Message{Message: "Ooof, nice try"}:
			logging.Debug("usr sent sent")
		case <-ctx.Done():
			logging.Debug("usr sent closed by context")
			return
		}

		// tr := transcriptaudio.New(scnr.stt, translated.Audio.Data)
	}

	select {
	case scnr.ai.Reqs() <- ai.ReqAI{Words: &Words}:
		logging.Debug("ai sent sent")

	case <-ctx.Done():
		logging.Debug("ai sent closed by context")
		return
	}

	// results := ai.RespAI{}
	select {
	case answ := <-scnr.ai.Resps():
		sentencesToTranslate = answ
		logging.Debug("ai got resp", "respId", answ, "respMsg", answ)

	case <-ctx.Done():
		logging.Debug("ai got resp")
		return
	}
}
