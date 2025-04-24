package model

type UserID uint64

type AudioData struct {
	Data []byte
}

type Message struct {
	Id      UserID
	Audio   *AudioData
	Message string
	Action  *Action
}

type Action struct {
	Type    string
	Message string
	Values  []string
}
