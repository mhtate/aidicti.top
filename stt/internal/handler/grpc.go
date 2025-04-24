package handler

import (
	"context"
	"fmt"

	"aidicti.top/api/protogen_cmn"
	"aidicti.top/api/protogen_stt"
	"aidicti.top/pkg/handler/exectime"
	"aidicti.top/pkg/handler/proc"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/stt/internal/model"
)

type controller interface {
	TranslateAudio(context.Context, pkgmodel.ReqData, *model.AudioData) (
		*model.TranscriptionResult, error)
}

type handlerProc_TranscriptAudio interface {
	Exec(context.Context, *protogen_stt.TranscriptAudioRequest) (
		*protogen_stt.TranscriptAudioResponse, error)
}

type handler struct {
	protogen_stt.UnimplementedSTTServiceServer
	ctrl controller

	procTranscriptAudio handlerProc_TranscriptAudio
}

func New(c controller) *handler {
	h := &handler{ctrl: c}

	h.procTranscriptAudio = exectime.New(
		proc.New(
			func(req *protogen_stt.TranscriptAudioRequest) (*model.AudioData, error) {

				if req.Data == nil {
					return nil, fmt.Errorf("req.Data != nil")
				}

				if len(req.Data.Data) == 0 {
					return nil, fmt.Errorf("len(req.Data.Data) != 0")
				}

				data := &model.AudioData{Data: req.Data.Data}

				return data, nil
			},

			func(resp *model.TranscriptionResult) (*protogen_stt.TranscriptAudioResponse, error) {
				return &protogen_stt.TranscriptAudioResponse{
					Id:            &protogen_cmn.ReqData{},
					Transcription: resp.Transcription}, nil
			},

			h.ctrl.TranslateAudio,
		))

	return h
}

func (h *handler) TranscriptAudio(ctx context.Context, req *protogen_stt.TranscriptAudioRequest) (
	*protogen_stt.TranscriptAudioResponse, error) {

	return h.procTranscriptAudio.Exec(ctx, req)
}
