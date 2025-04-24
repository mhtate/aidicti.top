package controller

import (
	"context"
	"fmt"
	"testing"

	"aidicti.top/ai/internal/model"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Valid words container returns expected sentence container
func TestGetSentencesReturnsExpectedContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockGateway := mocks.NewMockGateway(ctrl)

	dialog := &model.Dialog{}
	expectedSentences := model.SentenceContainer{
		Sentences: []string{"test sentence"},
	}

	mockStorage.EXPECT().Create().Return(dialog, nil)
	mockGateway.EXPECT().GetSentences(gomock.Any(), gomock.Any()).Return(expectedSentences, nil)

	c := NewController(mockStorage, mockGateway)

	words := model.WordsContainer{
		Words: []string{"test"},
	}

	result, err := c.GetSentences(context.Background(), words)

	assert.NoError(t, err)
	assert.Equal(t, expectedSentences, result)
}

// Empty words container triggers panic
func TestGetSentencesPanicsOnEmptyWords(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockGateway := mocks.NewMockGateway(ctrl)

	c := NewController(mockStorage, mockGateway)

	words := model.WordsContainer{
		Words: []string{},
	}

	assert.Panics(t, func() {
		c.GetSentences(context.Background(), words)
	})
}

// Storage Create() returns error
func TestGetSentencesStorageCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockGateway := mocks.NewMockGateway(ctrl)

	expectedError := fmt.Errorf("storage create error")

	mockStorage.EXPECT().Create().Return(nil, expectedError)

	c := NewController(mockStorage, mockGateway)

	words := model.WordsContainer{
		Words: []string{"test"},
	}

	_, err := c.GetSentences(context.Background(), words)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}
