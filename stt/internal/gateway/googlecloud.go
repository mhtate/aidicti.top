package gateway

import (
	"context"
	"fmt"

	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/utils"
	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
)

type STTRequest struct {
	Context context.Context
	Data    []byte
}

type STTResponse struct {
	Transcription string
	Confidence    float32
	Err           error
}

type gateway struct {
	client *speech.Client
}

func New() *gateway {
	//TODO make it canceled and cancel in Cancel method
	ctx := context.Background()
	client, err := speech.NewClient(ctx)
	utils.Assert(err == nil, fmt.Sprintf("Failed to start Google Cloud Speech Client [%s]", err))

	return &gateway{client: client}
}

func (g gateway) Close() {
	err := g.client.Close()

	utils.Assert(err == nil, "Error occured while closing Google Cloud Speech Client")
}

type Dialog interface {
	Requests() chan<- STTRequest
	Responses() <-chan STTResponse
	Done() chan<- struct{}
}

type dialog struct {
	gateway   *gateway
	requests  chan STTRequest
	responses chan STTResponse
	done      chan struct{}
}

func (d dialog) Requests() chan<- STTRequest {
	return d.requests
}
func (d dialog) Responses() <-chan STTResponse {
	return d.responses
}
func (d dialog) Done() chan<- struct{} {
	return d.done
}

func (g *gateway) NewDialog() Dialog {
	d := dialog{
		gateway:   g,
		requests:  make(chan STTRequest, 10),
		responses: make(chan STTResponse, 10),
		done:      make(chan struct{}),
	}

	go func() {
		for {
			select {
			case r := <-d.requests:
				resp, err := d.gateway.client.Recognize(r.Context, &speechpb.RecognizeRequest{
					Config: &speechpb.RecognitionConfig{
						Encoding:        speechpb.RecognitionConfig_OGG_OPUS,
						SampleRateHertz: 48000,
						LanguageCode:    "en-US",
					},
					Audio: &speechpb.RecognitionAudio{
						AudioSource: &speechpb.RecognitionAudio_Content{
							Content: r.Data,
						},
					},
				})

				if err != nil {
					logging.Info("Recognize failed", "err", err)
					d.responses <- STTResponse{Err: err}
					break
				}

				if len(resp.Results) == 0 {
					logging.Info("Recognize failed", "err", fmt.Errorf("There is no results"))
					d.responses <- STTResponse{Err: fmt.Errorf("There is no results")}
					break
				}

				if len(resp.Results[0].Alternatives) == 0 {
					logging.Info("Recognize failed", "err", fmt.Errorf("There is no results"))
					d.responses <- STTResponse{Err: fmt.Errorf("There is no results")}
					break
				}

				//The most probable alternative as doc SpeechRecognitionAlternative says
				tr := resp.Results[0].Alternatives[0]

				//TODO looks like we should return here
				d.responses <- STTResponse{tr.Transcript, tr.Confidence, nil}

			case <-d.done:
				return
			}
		}
	}()

	return d
}
