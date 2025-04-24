package model

type RequestID uint64
type UserID uint64
type DialogID uint64
type SessionID uint64

type AudioData struct {
	Data []byte
}

type Message struct {
	Id      UserID
	Audio   *AudioData
	Message string
	Action  *Action
	Actions []uint64
}

type Action struct {
	Type    string
	Message string
	Values  []string
}

type WordReq struct {
	Word string
	Info string
}

type WordsReq struct {
	Words []WordReq
}

type SentResp struct {
	Original    string
	Translation string
	Word        string
}

type SentsResp struct {
	Sents []SentResp
}

type TranslationResult struct {
	Correction   string
	Explaination string
	Rating       uint8
}

type TranslationResultContainer struct {
	Results []TranslationResult
}
