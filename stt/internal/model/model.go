package model

type RequestID uint64
type UserID uint64

type AudioData struct {
	Data []byte
}

type TranscriptionResult struct {
	Transcription string
	Confidence    float32
}
