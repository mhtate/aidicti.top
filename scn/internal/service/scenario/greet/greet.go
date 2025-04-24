package greet

import (
	"context"
	"strings"

	"aidicti.top/pkg/logging"
	"aidicti.top/scn/internal/model"
	"aidicti.top/scn/internal/service/scenario"
	"aidicti.top/scn/internal/service/scenario/createsentences"
	"aidicti.top/scn/internal/service/scenario/transcriptaudio"
)

type greet struct {
	usr scenario.DialogUSR
	stt scenario.DialogSTT
	ai  scenario.DialogGPT
}

func New(stt scenario.DialogSTT, usr scenario.DialogUSR, ai scenario.DialogGPT) *greet {
	return &greet{usr: usr, stt: stt, ai: ai}
}

// func (scnr greet) Run(ctx context.Context) {
// 	select {
// 	case scnr.usr.Reqs() <- model.Message{Message: "Hello"}:
// 		logging.Debug("greet sent")
// 	case <-ctx.Done():
// 		logging.Debug("greet closed by context")
// 		return
// 	}

// 	select {
// 	case answ := <-scnr.usr.Resps():
// 		logging.Debug("greet got resp", "resp", answ)
// 	case <-ctx.Done():
// 		logging.Debug("greet closed by context")
// 		return
// 	}

// 	logging.Debug("greet finished")
// }

func (scnr greet) Run(ctx context.Context) {
	// <-scnr.usr.Resps()

	select {
	case scnr.usr.Reqs() <- model.Message{Message: "Hello, please say several words for me to create sentences"}:
		logging.Debug("greet sent")
	case <-ctx.Done():
		logging.Debug("greet closed by context")
		return
	}
	for {
		select {
		case answ := <-scnr.usr.Resps():
			logging.Debug("greet got resp", "respId", answ.Id, "respMsg", answ.Message)

			scnr.usr.Reqs() <- model.Message{Message: "Yes, i got your message"}

			if answ.Audio != nil {
				tr := transcriptaudio.New(scnr.stt, answ.Audio.Data)

				tr.Run(ctx)

				scnr.usr.Reqs() <- model.Message{Message: "oh, i got your words"}

				words := strings.Fields(tr.Result)

				if len(words) < 4 {
					scnr.usr.Reqs() <- model.Message{Message: "oh, try one more time"}
					break
				}

				scnr.usr.Reqs() <- model.Message{Message: "it sounds like: " + tr.Result}

				cr := createsentences.New(scnr.ai, scnr.usr, words[0:4])

				cr.Run(ctx)

				return

			} else {
				scnr.usr.Reqs() <- model.Message{Message: "Hey, there is no audio :( try more"}
			}

		case <-ctx.Done():
			logging.Debug("greet closed by context")
			return
		}
		logging.Debug("greet finished")
	}

}
