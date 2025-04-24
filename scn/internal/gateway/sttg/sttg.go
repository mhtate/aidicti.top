package sttg

import (
	"context"

	"aidicti.top/api/protogen_cmn"
	"aidicti.top/api/protogen_stt"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/gatewaydialog/processor"
	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/utils"
	"aidicti.top/scn/internal/model"
)

type ReqSTT struct {
	ReqId     model.RequestID
	DialogID  model.DialogID
	AudioData model.AudioData
}

type RespSTT struct {
	ReqId    model.RequestID
	DialogID model.DialogID
	Data     string
}

type gateway struct {
	client    protogen_stt.STTServiceClient
	requests  chan ReqSTT
	responses chan RespSTT
}

func (g gateway) Reqs() chan<- ReqSTT {
	return g.requests
}

func (g gateway) Resps() <-chan RespSTT {
	return g.responses
}

func New(client protogen_stt.STTServiceClient) *gateway {
	return &gateway{
		client:    client,
		requests:  make(chan ReqSTT, 64),
		responses: make(chan RespSTT, 256),
	}
}

type reqProcessor struct {
	clnt  protogen_stt.STTServiceClient
	resps chan<- RespSTT
}

func (p reqProcessor) Process(ctx context.Context, req ReqSTT) error {
	utils.Assert(len(req.AudioData.Data) != 0, "invalid data size")

	data := &protogen_stt.AudioFileS{Data: req.AudioData.Data}

	resp, err := p.clnt.TranscriptAudio(ctx, &protogen_stt.TranscriptAudioRequest{
		Id: &protogen_cmn.ReqData{
			ReqId: uint64(req.ReqId),
			DlgId: uint64(req.ReqId),
		},
		RequestId: uint64(req.ReqId),
		UserId:    uint64(req.ReqId),
		Data:      data,
	})

	if err != nil {
		logging.Info("stt request error", "err", err)
		return err
	}

	p.resps <- RespSTT{
		ReqId:    model.RequestID(resp.Id.ReqId),
		DialogID: model.DialogID(resp.Id.DlgId),
		Data:     resp.Transcription,
	}

	logging.Debug("stt received resp", "resp", resp)

	return nil
}

func (g gateway) Run(ctx context.Context) {
	reqLstn := listener.New(
		g.requests,
		processor.New(reqProcessor{clnt: g.client, resps: g.responses}, 16))

	reqLstn.Run(ctx)
}
