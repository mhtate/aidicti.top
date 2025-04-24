package service_test

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"log/slog"

	"aidicti.top/oxf/internal/gateway/oxf"
	"aidicti.top/oxf/internal/model"
	"aidicti.top/oxf/internal/service"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"github.com/stretchr/testify/assert"
)

type gatewayMock struct {
	reqs  chan oxf.Req
	resps chan oxf.Resp
}

func (g gatewayMock) Reqs() chan<- oxf.Req {
	return g.reqs
}

func (g gatewayMock) Resps() <-chan oxf.Resp {
	return g.resps
}

func (g gatewayMock) Run(ctx context.Context) {
	reqPassedNum := 0

	for {
		select {
		case req := <-g.reqs:

			g.resps <- oxf.Resp{
				ReqData:         req.ReqData,
				DictionaryEntry: model.DictionaryEntry{Word: req.Text}}

			if reqPassedNum < 4 {

				reqPassedNum += 1
			}
		case <-ctx.Done():
			return
		}
	}
}

// type MockStorage struct {
// 	mock.Mock
// }

// func (m *MockStorage) Get(id pkgmodel.DlgID) (gatewaydialog.Dialog[model.Word, model.DictionaryEntry], error) {
// 	args := m.Called(id)
// 	return args.Get(0).(gatewaydialog.Dialog[model.Word, model.DictionaryEntry]), args.Error(1)
// }

// Mocking the dialog
// type MockDialog struct {
// 	reqs  chan model.Word
// 	resps chan model.DictionaryEntry
// }

// func (m *MockDialog) Reqs() chan<- model.Word {
// 	return m.reqs
// }

// func (m *MockDialog) Resps() <-chan model.DictionaryEntry {
// 	return m.resps
// }

func Test_GetDictEntry_Success(t *testing.T) {
	// logging.Set(slog.New(slog.NewJSONHandler(NullWriter{}, nil)))
	logging.Set(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	ctx, cancel := context.WithCancel(context.Background())

	gtw := gatewayMock{reqs: make(chan oxf.Req, 8), resps: make(chan oxf.Resp, 8)}
	go gtw.Run(ctx)

	s := service.New(gtw)
	go s.Run(ctx)

	time.Sleep(10 * time.Millisecond)

	e1, err := s.GetDictEntry(ctx, pkgmodel.ReqData{1, 1}, model.Word{Text: "hello"})
	assert.NoError(t, err)
	assert.Equal(t, e1.Word, "hello")

	e2, err := s.GetDictEntry(ctx, pkgmodel.ReqData{2, 2}, model.Word{Text: "world"})
	assert.NoError(t, err)
	assert.Equal(t, e2.Word, "world")

	e3, err := s.GetDictEntry(ctx, pkgmodel.ReqData{3, 3}, model.Word{Text: "foo"})
	assert.NoError(t, err)
	assert.Equal(t, e3.Word, "foo")

	e4, err := s.GetDictEntry(ctx, pkgmodel.ReqData{1, 1}, model.Word{Text: "bar"})
	assert.NoError(t, err)
	assert.Equal(t, e4.Word, "bar")

	makeReq := func() {
		ids := pkgmodel.ReqData{pkgmodel.ReqID(rand.Uint64()), pkgmodel.DlgID(rand.Uint64())}
		data := strconv.FormatUint(uint64(rand.Uint64()), 10)

		ans, err := s.GetDictEntry(ctx, ids, model.Word{Text: data})
		assert.NoError(t, err)
		assert.Equal(t, ans.Word, data)
	}

	for _ = range 200 {
		go makeReq()
	}

	makeReq()

	cancel()

	for _ = range 200 {
		go makeReq()
	}

	time.Sleep(2 * time.Second)
}

// func TestGetDictEntryDialogNotReceiving(t *testing.T) {
// 	ctx := context.Background()
// 	mockStorage := new(MockStorage)
// 	mockGateway := new(MockGateway)
// 	svc := &service{gtw: mockGateway, stg: mockStorage}

// 	dialog := &MockDialog{
// 		reqs:  make(chan model.Word),
// 		resps: make(chan model.DictionaryEntry),
// 	}

// 	mockStorage.On("Get", mock.Anything).Return(dialog, nil)

// 	ids := pkgmodel.ReqData{DlgId: "valid-id"}
// 	word := model.Word("test")

// 	entry, err := svc.GetDictEntry(ctx, ids, word)
// 	assert.NoError(t, err)
// 	assert.Equal(t, model.DictionaryEntry{}, entry)
// }

// func TestGetDictEntryContextDone(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	cancel() // Immediately cancel the context

// 	mockStorage := new(MockStorage)
// 	mockGateway := new(MockGateway)
// 	svc := &service{gtw: mockGateway, stg: mockStorage}

// 	dialog := &MockDialog{
// 		reqs:  make(chan model.Word, 1),
// 		resps: make(chan model.DictionaryEntry),
// 	}

// 	mockStorage.On("Get", mock.Anything).Return(dialog, nil)

// 	ids := pkgmodel.ReqData{DlgId: "valid-id"}
// 	word := model.Word("test")

// 	entry, err := svc.GetDictEntry(ctx, ids, word)
// 	assert.Error(t, err)
// 	assert.Equal(t, model.DictionaryEntry{}, entry)
// }
