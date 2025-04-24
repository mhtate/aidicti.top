package model

// TODO to DialogId
type DlgID uint64
type ReqID uint64
type DialogID uint64

type ReqData struct {
	ReqId ReqID
	DlgId DlgID
}

const (
	ServiceSCN = "scnr"
	ServiceTLG = "tlg"
	ServiceGPT = "gpt"
	ServiceOXF = "oxf"
	ServiceSTT = "stt"
	ServiceUIS = "uis"
)
