package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"aidicti.top/pkg/logging"
	"aidicti.top/stt/internal/gateway"
	"aidicti.top/stt/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestTranslateAudio_Success(t *testing.T) {
	logging.Set(slog.New(slog.NewJSONHandler(logging.NullWriter{}, nil)))

	gateway := gateway.New()
	defer gateway.Close()

	service := NewService(gateway)
	ctx := context.Background()

	data, err := os.ReadFile("testdata/audio1.ogg")
	if err != nil {
		assert.Fail(t, "err not nil")
	}

	result, err := service.TranslateAudio(ctx,
		model.RequestID(42),
		model.UserID(24),
		model.AudioData{Data: data})
	if err != nil {
		assert.Fail(t, "err not nil")
	}

	if result.Confidence < 0.7 {
		assert.Fail(t, "result.Confidence > 0.7")
	}

	assert.Equal(t, result.Transcription, "hello my dear friends it's a voice message", "")
}
