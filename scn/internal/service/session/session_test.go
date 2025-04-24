package session_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/model/oxf"
	"aidicti.top/scenario/internal/gateway/ai"
	"aidicti.top/scenario/internal/model"
	"aidicti.top/scenario/internal/service/scenario"
	"aidicti.top/scenario/internal/service/scenario/getdictentry"
	"aidicti.top/scenario/internal/service/session"
	"github.com/stretchr/testify/assert"
)

type dialogSTT struct {
	Reqs_  chan model.AudioData
	Resps_ chan string
}

func (d dialogSTT) Reqs() chan model.AudioData { return d.Reqs_ }
func (d dialogSTT) Resps() chan string         { return d.Resps_ }

type dialogUSR struct {
	Reqs_  chan model.Message
	Resps_ chan model.Message
}

func (d dialogUSR) Reqs() chan model.Message  { return d.Reqs_ }
func (d dialogUSR) Resps() chan model.Message { return d.Resps_ }

type dialogGPT struct {
	Reqs_  chan ai.ReqAI
	Resps_ chan ai.RespAI
}

func (d dialogGPT) Reqs() chan ai.ReqAI   { return d.Reqs_ }
func (d dialogGPT) Resps() chan ai.RespAI { return d.Resps_ }

type dialogOXF struct {
	Reqs_  chan oxf.Word
	Resps_ chan oxf.DictionaryEntry
}

func (d dialogOXF) Reqs() chan oxf.Word             { return d.Reqs_ }
func (d dialogOXF) Resps() chan oxf.DictionaryEntry { return d.Resps_ }

// Session correctly handles user responses and forwards them to the scenario
func Test_Session_Set_Scnr(t *testing.T) {
	// logging.Set(slog.New(slog.NewJSONHandler(NullWriter{}, nil)))
	logging.Set(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	usr := dialogUSR{Reqs_: make(chan model.Message), Resps_: make(chan model.Message)}
	gpt := dialogGPT{Reqs_: make(chan ai.ReqAI), Resps_: make(chan ai.RespAI)}
	stt := dialogSTT{Reqs_: make(chan model.AudioData), Resps_: make(chan string)}
	oxf := dialogOXF{Reqs_: make(chan oxf.Word), Resps_: make(chan oxf.DictionaryEntry)}

	s := session.New(stt, usr, gpt, oxf)

	s.Set(
		func(
			stt scenario.DialogSTT,
			usr scenario.DialogUSR,
			gpt scenario.DialogGPT,
			oxf scenario.DialogOXF) scenario.Scenario {

			return getdictentry.New(stt, usr, gpt, oxf)
		})

	go s.Run(ctx)

	req := model.Message{}

	req = <-usr.Reqs()
	assert.Equal(t, req.Message, "Please write a word")

	usr.Resps() <- model.Message{Message: ""}

	req = <-usr.Reqs()
	assert.Equal(t, req.Message, "Please write something")

	requestWord := "hello"
	usr.Resps() <- model.Message{Message: requestWord}

	reqOXF := <-oxf.Reqs()
	assert.Equal(t, requestWord, reqOXF.Text)

}

// // Context cancellation properly terminates the session
// func TestSessionTerminatesOnContextCancellation(t *testing.T) {
// 	// Create a context with cancel function
// 	ctx, cancel := context.WithCancel(context.Background())

// 	// Create a channel to track if the function returns
// 	done := make(chan struct{})

// 	// Create mock channels
// 	mockUser := &mockResponder{resps: make(chan interface{}, 1)}
// 	mockScenarioUser := &mockResponder{resps: make(chan interface{}, 1)}
// 	mockGPT := &mockResponder{resps: make(chan interface{}, 1)}
// 	mockSTT := &mockResponder{resps: make(chan interface{}, 1)}
// 	mockOXF := &mockResponder{resps: make(chan interface{}, 1)}
// 	mockScenarioGPT := &mockResponder{resps: make(chan interface{}, 1)}
// 	mockScenarioSTT := &mockResponder{resps: make(chan interface{}, 1)}
// 	mockScenarioOXF := &mockResponder{resps: make(chan interface{}, 1)}

// 	// Create a mock cancel function to verify it's called
// 	cancelCalled := false
// 	mockCancel := func() {
// 		cancelCalled = true
// 	}

// 	// Create session
// 	s := &session{
// 		id:      uuid.New().String(),
// 		usr:     mockUser,
// 		scnrUSR: mockScenarioUser,
// 		gpt:     mockGPT,
// 		scnrGPT: mockScenarioGPT,
// 		stt:     mockSTT,
// 		scnrSTT: mockScenarioSTT,
// 		oxf:     mockOXF,
// 		scnrOXF: mockScenarioOXF,
// 		mtx:     &sync.Mutex{},
// 		cancel:  mockCancel,
// 	}

// 	// Run the session in a goroutine
// 	go func() {
// 		s.Run(ctx)
// 		close(done)
// 	}()

// 	// Cancel the context
// 	cancel()

// 	// Wait for the function to return or timeout
// 	select {
// 	case <-done:
// 		// Success - function returned
// 	case <-time.After(time.Second):
// 		t.Fatal("Session did not terminate after context cancellation")
// 	}

// 	// Verify the session's cancel function was called
// 	assert.True(t, cancelCalled, "Session's cancel function was not called")
// }
