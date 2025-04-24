package service

import (
	"context"
	"fmt"
	"time"

	"aidicti.top/pkg/logging"
	"aidicti.top/stt/internal/gateway"
	"aidicti.top/stt/internal/model"
)

type STTGateway interface {
	NewDialog() gateway.Dialog
}

type service struct {
	channels map[model.UserID](gateway.Dialog)
	gateway  STTGateway
}

func NewService(g STTGateway) service {
	return service{channels: make(map[model.UserID]gateway.Dialog), gateway: g}
}

func (s service) TranslateAudio(ctx context.Context, requestId model.RequestID, userId model.UserID, audioData model.AudioData) (
	model.TranscriptionResult, error) {
	dialog := s.gateway.NewDialog()

	cancelCtx, cancelF := context.WithCancel(ctx)

	_ = cancelF

	select {
	case dialog.Requests() <- gateway.STTRequest{
		Context: cancelCtx,
		Data:    audioData.Data,
	}:
		logging.Info("passed request")
	default:
		logging.Info("dialog didnt get request")
	}

	// cancelF()

	select {
	case response := <-dialog.Responses():
		logging.Info("Response recieved")
		if response.Err != nil {
			logging.Info("Response got error", response.Err)
			close(dialog.Done())
			return model.TranscriptionResult{}, fmt.Errorf("task is finished with an error")
		}

		logging.Info("Sending response", "transcription", response.Transcription, "Confidence", response.Confidence)
		return model.TranscriptionResult{
			Transcription: response.Transcription,
			Confidence:    response.Confidence,
		}, nil

	case <-ctx.Done():
		logging.Info("Context Done")
		close(dialog.Done())
		return model.TranscriptionResult{}, fmt.Errorf("task is cancelled")

	case <-time.After(10 * time.Second):
		logging.Info("Timeout")
		close(dialog.Done())
		return model.TranscriptionResult{}, fmt.Errorf("task is cancelled")
	}
}
