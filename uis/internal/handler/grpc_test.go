package handler_test

// MockController is a mock implementation of the controller interface
// type MockController struct {
// 	mock.Mock
// }

// func (m *MockController) GetDictEntry(ctx context.Context, ids model.ReqData, word model.Word) (model.DictionaryEntry, error) {
// 	args := m.Called(ctx, ids, word)
// 	return args.Get(0).(model.DictionaryEntry), args.Error(1)
// }

// func TestGetDictEntry(t *testing.T) {
// 	mockCtrl := new(MockController)
// 	h := New(mockCtrl)

// 	testWord := &oxf_gen_proto.Word{Text: "example", Id: "1", UserId: "user1"}
// 	expectedEntry := model.DictionaryEntry{
// 		Word:         "example",
// 		PartOfSpeech: "noun",
// 	}

// 	mockCtrl.On("GetDictEntry", mock.Anything, mock.Anything, model.Word{Text: "example"}).Return(expectedEntry, nil)

// 	resp, err := h.GetDictEntry(context.Background(), testWord)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, resp)
// 	assert.Equal(t, "example", resp.Word)
// 	assert.Equal(t, "noun", resp.PartOfSpeech)
// }

// func BenchmarkGetDictEntry(b *testing.B) {
// 	mockCtrl := new(MockController)
// 	h := New(mockCtrl)

// 	testWord := &oxf_gen_proto.Word{Text: "example", Id: "1", UserId: "user1"}
// 	expectedEntry := model.DictionaryEntry{
// 		Word:         "example",
// 		PartOfSpeech: "noun",
// 	}

// 	mockCtrl.On("GetDictEntry", mock.Anything, mock.Anything, model.Word{Text: "example"}).Return(expectedEntry, nil)

// 	for i := 0; i < b.N; i++ {
// 		_, _ = h.GetDictEntry(context.Background(), testWord)
// 	}
// }
