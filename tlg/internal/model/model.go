package model

type RequestID uint64
type UserID uint64

type AudioData struct {
	Data []byte
}

type Message struct {
	Id      UserID //we should delete all ids from model innit?
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
