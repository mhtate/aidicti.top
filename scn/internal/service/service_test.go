package service

import (
	"context"
	"os"
	"testing"

	"aidicti.top/stt/internal/gateway"
	"aidicti.top/stt/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestTranslateAudio_Success(t *testing.T) {
	gateway := gateway.New()
	defer gateway.Close()

	service := NewService(gateway)
	ctx := context.Background()

	data, err := os.ReadFile("testdata/audio1.ogg")
	if err != nil {
		assert.Fail(t, "err not nil")
	}

	result, err := service.TranslateAudio(ctx, model.AudioContainer{Data: data})
	if err != nil {
		assert.Fail(t, "err not nil")
	}

	if result.Confidence < 0.7 {
		assert.Fail(t, "result.Confidence > 0.7")
	}

	assert.Equal(t, result.Transcription, "hello my dear friends it's a voice message", "")
}
