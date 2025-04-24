package transcriptaudio

import (
	"context"
	"time"

	"aidicti.top/pkg/logging"
	"aidicti.top/scn/internal/model"
	"aidicti.top/scn/internal/service/scenario"
)

type transcriptAudio struct {
	audiofile []byte
	sttd      scenario.DialogSTT
	Result    string
}

func New(sttd scenario.DialogSTT, audiofile []byte) *transcriptAudio {
	return &transcriptAudio{audiofile: audiofile, sttd: sttd}
}

func (scnr *transcriptAudio) Run(ctx context.Context) {
	select {
	case scnr.sttd.Reqs() <- model.AudioData{Data: scnr.audiofile}:
		logging.Debug("transcriptAudio sent")
	case <-ctx.Done():
		logging.Debug("transcriptAudio closed by context")
		return
	}

	select {
	case answ := <-scnr.sttd.Resps():
		logging.Debug("transcriptAudio got resp", "resp", answ)

		scnr.Result = answ
	case <-ctx.Done():
		logging.Debug("transcriptAudio closed by context")
		return
	case <-time.After(3000 * time.Millisecond):
		logging.Debug("transcriptAudio closed by timeout")
		return
	}

	logging.Debug("transcriptAudio finished")
}
