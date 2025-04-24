package oxf_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"aidicti.top/oxf/internal/gateway/oxf"
	"aidicti.top/oxf/internal/model"
	"aidicti.top/pkg/gatewaydialog"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockServiceOXFClient is a mock implementation of the ServiceOXFClient interface
type MockServiceOXFClient struct {
	mock.Mock
}

func (m *MockServiceOXFClient) SomeClientMethod() {
	// Mock method implementation
}

// MockParse is a mock implementation of the parse package
var MockParse = struct {
	GetEntry func(html string) (*model.DictionaryEntry, error)
}{
	GetEntry: func(html string) (*model.DictionaryEntry, error) {
		return &model.DictionaryEntry{}, nil
	},
}

// MockFetchHTML is a mock implementation of the fetchHTML function
var MockFetchHTML = func(url string) (string, error) {
	return "<html></html>", nil
}

func Test_Gateway_Success(t *testing.T) {
	logging.Set(slog.New(slog.NewJSONHandler(logging.NullWriter{}, nil)))

	ctx := context.Background()

	gtw := oxf.New()
	gtw.SetProcessor(
		func(p gatewaydialog.Processor[oxf.Req]) gatewaydialog.Processor[oxf.Req] {
			return p
		})
	go gtw.Run(ctx)

	req := oxf.Req{
		ReqData: pkgmodel.ReqData{
			ReqId: 1,
			DlgId: 1,
		},
		Word: model.Word{Text: "just"},
	}

	gtw.Reqs() <- req

	select {
	case resp := <-gtw.Resps():
		assert.Equal(t, "www.oxfordlearnersdictionaries.com/definition/english/just_1", resp.DictionaryEntry.Link)
	case <-time.After(5 * time.Second):
		assert.Fail(t, "out by timeout")
	}
}

// func TestProcess_FetchHTMLFails(t *testing.T) {
// 	client := &MockServiceOXFClient{}
// 	resps := make(chan oxf.RespOXF, 1)
// 	processor := oxf.New(client)

// 	// Override fetchHTML to simulate failure
// 	oxf.FetchHTML = func(url string) (string, error) {
// 		return "", errors.New("network error")
// 	}

// 	req := oxf.ReqOXF{
// 		ReqData: model.ReqData{
// 			ReqId: "test-req-id",
// 			DlgId: "test-dlg-id",
// 		},
// 		Word: "test",
// 	}

// 	go processor.Process(context.Background(), req)

// 	// Since fetchHTML fails, no response should be sent
// 	select {
// 	case <-resps:
// 		t.Fatal("Expected no response due to fetchHTML failure")
// 	default:
// 	}
// }

// func TestProcess_EmptyWord(t *testing.T) {
// 	client := &MockServiceOXFClient{}
// 	resps := make(chan oxf.RespOXF, 1)
// 	processor := oxf.New(client)

// 	req := oxf.ReqOXF{
// 		ReqData: model.ReqData{
// 			ReqId: "test-req-id",
// 			DlgId: "test-dlg-id",
// 		},
// 		Word: "",
// 	}

// 	assert.Panics(t, func() {
// 		processor.Process(context.Background(), req)
// 	}, "Expected panic due to empty word")
// }
