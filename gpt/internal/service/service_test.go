package service

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"aidicti.top/gpt/internal/gateway"
	"aidicti.top/gpt/internal/model"
	"aidicti.top/pkg/utils"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

// // MockAIGateway is a mock implementation of the AIGateway interface
// type MockAIGateway struct {
// 	mock.Mock
// }

// func (m *MockAIGateway) NewDialog() gateway.Dialog {
// 	args := m.Called()
// 	return args.Get(0).(gateway.Dialog)
// }

// // MockDialog is a mock implementation of the gateway.Dialog interface
// type MockDialog struct {
// 	mock.Mock
// 	requests  chan gateway.AIRequest
// 	responses chan gateway.AIResponse
// 	done      chan struct{}
// }

// func (m *MockDialog) Requests() chan<- gateway.AIRequest {
// 	return m.requests
// }

// func (m *MockDialog) Responses() <-chan gateway.AIResponse {
// 	return m.responses
// }

// func (m *MockDialog) Done() chan<- struct{} {
// 	return m.done
// }

// func TestGetSentences_Success(t *testing.T) {
// 	// Arrange
// 	mockGateway := new(MockAIGateway)
// 	mockDialog := &MockDialog{
// 		requests:  make(chan gateway.AIRequest, 1),
// 		responses: make(chan gateway.AIResponse, 1),
// 		done:      make(chan struct{}),
// 	}

// 	mockGateway.On("NewDialog").Return(mockDialog)

// 	service := NewService(mockGateway)

// 	ctx := context.Background()
// 	wordsContainer := model.WordsContainer{
// 		ID:    "test-id",
// 		Words: []model.Word{{Word: "test", Info: "info"}},
// 	}

// 	expectedResponse := gateway.AIResponse{
// 		C: `{"sentences": [{"id": 1, "original_word_required": "test", "sentence": "This is a test sentence."}]}`,
// 	}

// 	mockDialog.responses <- expectedResponse

// 	// Act
// 	result, err := service.GetSentences(ctx, wordsContainer)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.Equal(t, "test-id", result.ID)
// 	assert.Len(t, result.Sentences, 1)
// 	assert.Equal(t, "This is a test sentence.", result.Sentences[0].Sentence)
// 	assert.Equal(t, "test", result.Sentences[0].OriginalWord)

// 	mockGateway.AssertExpectations(t)
// }

// func TestGetSentences_Error(t *testing.T) {
// 	// Arrange
// 	mockGateway := new(MockAIGateway)
// 	mockDialog := &MockDialog{
// 		requests:  make(chan gateway.AIRequest, 1),
// 		responses: make(chan gateway.AIResponse, 1),
// 		done:      make(chan struct{}),
// 	}

// 	mockGateway.On("NewDialog").Return(mockDialog)

// 	service := NewService(mockGateway)

// 	ctx := context.Background()
// 	wordsContainer := model.WordsContainer{
// 		ID:    "test-id",
// 		Words: []model.Word{{Word: "test", Info: "info"}},
// 	}

// 	mockDialog.responses <- gateway.AIResponse{Err: errors.New("AI error")}

// 	// Act
// 	result, err := service.GetSentences(ctx, wordsContainer)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Equal(t, "task is finished with an error", err.Error())
// 	assert.Empty(t, result.Sentences)

// 	mockGateway.AssertExpectations(t)
// }

// func TestCheckTranslations_Success(t *testing.T) {
// 	// Arrange
// 	mockGateway := new(MockAIGateway)
// 	mockDialog := &MockDialog{
// 		requests:  make(chan gateway.AIRequest, 1),
// 		responses: make(chan gateway.AIResponse, 1),
// 		done:      make(chan struct{}),
// 	}

// 	mockGateway.On("NewDialog").Return(mockDialog)

// 	service := NewService(mockGateway)
// 	service.channels["test-id"] = mockDialog

// 	ctx := context.Background()
// 	translatedContainer := model.TranslatedSentenceContainer{
// 		ID:        "test-id",
// 		Sentences: []model.TranslatedSentence{{Sentence: "This is a test sentence."}},
// 	}

// 	expectedResponse := gateway.AIResponse{
// 		C: `{"sentences": [{"id": 1, "correction": "corrected sentence", "explanation": "explanation", "rating": 5}]}`,
// 	}

// 	mockDialog.responses <- expectedResponse

// 	// Act
// 	result, err := service.CheckTranslations(ctx, translatedContainer)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.NotNil(t, result)

// 	mockGateway.AssertExpectations(t)
// }

// func TestCheckTranslations_Error(t *testing.T) {
// 	// Arrange
// 	mockGateway := new(MockAIGateway)
// 	mockDialog := &MockDialog{
// 		requests:  make(chan gateway.AIRequest, 1),
// 		responses: make(chan gateway.AIResponse, 1),
// 		done:      make(chan struct{}),
// 	}

// 	mockGateway.On("NewDialog").Return(mockDialog)

// 	service := NewService(mockGateway)
// 	service.channels["test-id"] = mockDialog

// 	ctx := context.Background()
// 	translatedContainer := model.TranslatedSentenceContainer{
// 		ID:        "test-id",
// 		Sentences: []model.TranslatedSentence{{Sentence: "This is a test sentence."}},
// 	}

// 	mockDialog.responses <- gateway.AIResponse{Err: errors.New("AI error")}

// 	// Act
// 	result, err := service.CheckTranslations(ctx, translatedContainer)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Equal(t, "task is finished with an error", err.Error())
// 	assert.Empty(t, result)

// 	mockGateway.AssertExpectations(t)
// }

// func TestCheckTranslations_DialogNotFound(t *testing.T) {
// 	// Arrange
// 	mockGateway := new(MockAIGateway)
// 	service := NewService(mockGateway)

// 	ctx := context.Background()
// 	translatedContainer := model.TranslatedSentenceContainer{
// 		ID:        "test-id",
// 		Sentences: []model.TranslatedSentence{{Sentence: "This is a test sentence."}},
// 	}

// 	// Act
// 	result, err := service.CheckTranslations(ctx, translatedContainer)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Equal(t, "dialog is not found", err.Error())
// 	assert.Empty(t, result)

// 	mockGateway.AssertExpectations(t)
// }

type InjectionGateway struct {
	gateway AIGateway
}

type InjectionDialog struct {
	realDialog  gateway.Dialog
	callCounter uint
	requests    chan gateway.AIRequest
	responces   chan gateway.AIResponse
	done        chan struct{}
}

func (d InjectionDialog) Requests() chan<- gateway.AIRequest {
	return d.requests
}

func (d InjectionDialog) Responses() <-chan gateway.AIResponse {
	return d.responces
}

func (d InjectionDialog) Done() chan<- struct{} {
	return d.realDialog.Done()
}

type Dialog interface {
	Requests() chan<- gateway.AIRequest
	Responses() <-chan gateway.AIResponse
	Done() chan<- struct{}
}

type InjectionFailureGateway struct {
	gateway AIGateway
}

type InjectionFailureDialog struct {
	realDialog  gateway.Dialog
	callCounter uint
	requests    chan gateway.AIRequest
	responces   chan gateway.AIResponse
	done        chan struct{}
}

func (d InjectionFailureDialog) Requests() chan<- gateway.AIRequest {
	return d.requests
}

func (d InjectionFailureDialog) Responses() <-chan gateway.AIResponse {
	return d.responces
}

func (d InjectionFailureDialog) Done() chan<- struct{} {
	return d.done
}

func (g *InjectionFailureGateway) NewDialog() gateway.Dialog {
	realDialog := g.gateway.NewDialog()

	d := InjectionFailureDialog{
		realDialog:  realDialog,
		callCounter: 0,
		requests:    make(chan gateway.AIRequest),
		responces:   make(chan gateway.AIResponse),
		done:        make(chan struct{}),
	}

	go func() {
		for {
			select {
			case r := <-d.requests:
				if d.callCounter == 0 {
					d.callCounter += 1

					_ = r

					break
				}

			case r := <-d.realDialog.Responses():
				d.responces <- r

			case <-d.done:
				close(d.realDialog.Done())
				return
			}
		}
	}()

	return d
}

func (g *InjectionGateway) NewDialog() gateway.Dialog {
	realDialog := g.gateway.NewDialog()

	d := InjectionDialog{
		realDialog:  realDialog,
		callCounter: 0,
		requests:    make(chan gateway.AIRequest),
		responces:   make(chan gateway.AIResponse),
		done:        make(chan struct{}),
	}

	go func() {
		for {
			select {
			case r := <-d.requests:
				//if we were already called it means we already are injected
				if d.callCounter > 0 {
					d.realDialog.Requests() <- r
					break
				}

				//it's not thread safe but we will call it from one goroutine
				d.callCounter += 1

				dialogValue := reflect.ValueOf(d.realDialog)
				messagesField := dialogValue.FieldByName("Messages")
				if !messagesField.IsValid() {
					panic("messagesField should be valid")
				}
				messages := messagesField.Interface().([]openai.ChatCompletionMessage)
				utils.Assert(messages != nil, "messages must be valid")

				//put request (GetSentence)
				messages = append(messages,
					openai.ChatCompletionMessage{
						Role:    roleToString(r.Role),
						Content: string(r.Content),
					})

				name := "set_sentences"
				args := "{\"sentences\": [{\"id\": 1, \"original_word_required\": \"break up\", \"sentence\": \"Они решили расстаться после 5 лет совместной жизни\"}, {\"id\": 2, \"original_word_required\": \"a rule of thumb\", \"sentence\": \"Хорошее практическое правило — всегда перепроверять свою работу.\"}, {\"id\": 3, \"original_word_required\": \"rusty\", \"sentence\": \"Я чувствую себя усталым от игры в теннис после того, как некоторое время не играл.\"}]}"

				//put answer (GetSentence)
				messages = append(messages,
					openai.ChatCompletionMessage{
						Role:    roleToString(r.Role),
						Content: "",
						ToolCalls: []openai.ToolCall{
							openai.ToolCall{
								ID:   "call_i1blsxlG9mjWdv6mCQhHH1FJ",
								Type: "function",
								Function: openai.FunctionCall{
									Name:      name,
									Arguments: args,
								},
							},
						},
					})

				d.responces <- gateway.AIResponse{
					Content: "",
					C:       model.Call(args),
					Err:     nil,
				}
			case r := <-d.realDialog.Responses():
				d.responces <- r

			case <-d.done:
				close(d.realDialog.Done())
				return
			}
		}
	}()

	return d
}

func TestCheckTranslations_Injection_Success(t *testing.T) {
	gateway := InjectionGateway{gateway: gateway.New()}
	service := NewService(&gateway)
	ctx := context.Background()

	_, err := service.GetSentences(ctx, model.WordsContainer{})
	if err != nil {
		assert.Fail(t, "err not nil")
	}

	check, err := service.CheckTranslations(ctx, model.TranslatedSentenceContainer{
		Sentences: []model.TranslatedSentence{
			model.TranslatedSentence{Sentence: "They leave house apart before 5 years of hard working"},
			model.TranslatedSentence{Sentence: "A good rule of thumb is to always double-check your work."},
			model.TranslatedSentence{Sentence: "I'm feeling rusty at playing tennis after not playing for a while."},
		},
	})

	if err != nil {
		assert.Fail(t, err.Error())
	}

	if len(check.Results) != 3 {
		assert.Fail(t, "incorrect count of results")
	}

	const MaxRating = 5

	if check.Results[0].Rating == MaxRating {
		assert.Fail(t, "incorrect rating")
	}

	if check.Results[1].Rating != MaxRating {
		assert.Fail(t, "incorrect rating")
	}

	if check.Results[2].Rating == MaxRating {
		assert.Fail(t, "incorrect rating")
	}
}

func TestCheckTranslations_Injection_Failure(t *testing.T) {

}

func TestGetSentences_Success(t *testing.T) {

	gtw := gateway.New()

	srvc := NewService(gtw)

	ctx := context.Background()

	words := model.WordsContainer{
		ID: 42,
		Words: []model.Word{
			model.Word{
				ID:   421,
				Word: "come up (to somebody)",
				Info: "phrasal verb, to move towards somebody, in order to talk to them",
			},

			model.Word{
				ID:   422,
				Word: "power",
				Info: "a particular ability of the body or mind",
			},

			model.Word{
				ID:   423,
				Word: "justify",
				Info: "verb, justify somebody/something doing something, to give an explanation or excuse for something or for doing something",
			},

			model.Word{
				ID:   424,
				Word: "sell yourself (to somebody)",
				Info: "to accept money or a reward from somebody for doing something that is against your principles",
			},
		},
	}

	sntc, err := srvc.GetSentences(ctx, words)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(sntc.String())
	fmt.Println(sntc)

}
