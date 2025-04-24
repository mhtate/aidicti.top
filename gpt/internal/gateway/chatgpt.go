package gateway

import (
	"context"
	"encoding/json"
	"os"

	"aidicti.top/gpt/internal/model"
	"aidicti.top/pkg/utils"
	"github.com/sashabaranov/go-openai"
)

type AIRequest struct {
	Context context.Context
	Content model.Content
	F       *model.Func
	Role    model.Role
}

type AIResponse struct {
	Content model.Content
	C       model.Call
	Err     error
}

type Dialog interface {
	Requests() chan<- AIRequest
	Responses() <-chan AIResponse
	Done() chan<- struct{}
}

type privateKey struct {
	Key string `json:"private_key"`
}

type gateway struct {
	client *openai.Client
}

type dialog struct {
	//TODO this means we need more up entity with dialogs
	Messages  []openai.ChatCompletionMessage
	gateway   *gateway
	requests  chan AIRequest
	responses chan AIResponse
	done      chan struct{}
}

func (d dialog) Requests() chan<- AIRequest {
	return d.requests
}
func (d dialog) Responses() <-chan AIResponse {
	return d.responses
}

func (d dialog) Reqs() chan<- AIRequest {
	return d.requests
}
func (d dialog) Resps() <-chan AIResponse {
	return d.responses
}

func (d dialog) Done() chan<- struct{} {
	return d.done
}

func New() *gateway {
	const OpenAICredPathEnv = "OPENAI_APPLICATION_CREDENTIALS"
	value := os.Getenv(OpenAICredPathEnv)
	utils.Assert(value != "", "OPENAI_APPLICATION_CREDENTIALS env variable not set")

	file, err := os.Open(value)
	utils.Assert(err == nil, "OPENAI_APPLICATION_CREDENTIALS path got read with an error")
	defer file.Close()

	var privateKey privateKey
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&privateKey)
	utils.Assert(err == nil, "OPENAI_APPLICATION_CREDENTIALS json file got read with an error")

	return &gateway{client: openai.NewClient(privateKey.Key)}
}

func (g *gateway) NewDialog() Dialog {
	d := dialog{
		Messages:  make([]openai.ChatCompletionMessage, 0),
		gateway:   g,
		requests:  make(chan AIRequest),
		responses: make(chan AIResponse),
		done:      make(chan struct{}),
	}

	go func() {
		//TODO sure it's not closed yet
		defer close(d.responses)
		defer close(d.requests)

		for {
			select {
			case r := <-d.requests:
				d.Messages = append(d.Messages,
					openai.ChatCompletionMessage{
						Role:    roleToString(r.Role),
						Content: string(r.Content),
					})

				content, call, error := d.gateway.request(r.Context, &d, r.F)

				d.responses <- AIResponse{content, call, error}
			case <-d.done:
				return
			}
		}
	}()

	return d
}

func roleToString(role model.Role) string {
	switch role {
	case model.RoleSystem:
		return "system"
	case model.RoleUser:
		return "user"
	case model.RoleAssistant:
		return "assistant"
	case model.RoleFunction:
		return "function"
	case model.RoleTool:
		return "tool"
	default:
		return "unknown"
	}
}

func (d *dialog) Request(ctx context.Context, f *model.Func) (
	model.Content, model.Call, error) {

	return d.gateway.request(ctx, d, f)
}

// TODO fix a crunch with any in dialog
func (p *gateway) request(ctx context.Context, d *dialog, f *model.Func) (
	model.Content, model.Call, error) {

	utils.Assert(d != nil, "dialog is nil")

	tools := []openai.Tool{}
	if f != nil {
		tools = append(tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        f.Name,
				Description: f.Parameters.Description,
				Parameters:  f.Parameters,
			},
		})
	}

	const ai_model = openai.GPT4oMini
	resp, err := p.client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    ai_model,
			Messages: d.Messages,
			Tools:    tools,
		},
	)

	if err != nil {
		return "", "", err
	}

	responseMessage := resp.Choices[0].Message

	d.Messages = append(d.Messages, responseMessage)

	if len(responseMessage.ToolCalls) == 0 {
		return model.Content(responseMessage.Content), "", nil
	}

	return model.Content(responseMessage.Content), model.Call(responseMessage.ToolCalls[0].Function.Arguments), nil
}
