package gpt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"aidicti.top/gpt/internal/model"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/gatewaydialog/processor"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type ReqType = uint

const (
	GetSents   = ReqType(1)
	CheckSents = ReqType(2)
)

type Req struct {
	pkgmodel.ReqData
	Context context.Context
	Content model.Content
	F       *model.Func
	Role    model.Role
	//NOTE not so sure about that
	Meta any
	Tp   ReqType
}

type Resp struct {
	pkgmodel.ReqData
	Content model.Content
	C       model.Call
	Err     error
	Meta    any
}

type gateway struct {
	clnt  *openai.Client
	reqs  chan Req
	resps chan Resp
}

func (g gateway) Reqs() chan<- Req {
	return g.reqs
}

func (g gateway) Resps() <-chan Resp {
	return g.resps
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

	return &gateway{
		clnt:  openai.NewClient(privateKey.Key),
		reqs:  make(chan Req, 32),
		resps: make(chan Resp, 32),
	}
}

type reqProcessor struct {
	resps chan<- Resp
	clnt  *openai.Client
}

type fn struct {
	Name        string
	Description string
	Parameters  jsonschema.Definition
}

const originalWord = "originalWord"
const id = "id"
const sentence = "idealTranslationInEnglish"
const russian_sentence = "sentence"
const sentences = "sentences"

func getJsonFuncSchema() *jsonschema.Definition {
	schema := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			sentences: {
				Type: jsonschema.Array,
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						// id: {
						// 	Type:        jsonschema.Integer,
						// 	Description: id,
						// },
						originalWord: {
							Type:        jsonschema.String,
							Description: originalWord,
						},
						sentence: {
							Type:        jsonschema.String,
							Description: sentence,
						},
						russian_sentence: {
							Type:        jsonschema.String,
							Description: russian_sentence,
						},
					},
					Required: []string{id, originalWord, sentence, russian_sentence},
				},
			},
		},
		Required: []string{sentences},
	}

	return &schema
}

func (p reqProcessor) Process(ctx context.Context, req Req) error {

	if req.Tp == GetSents {

		//NOTE we don't need prev answers that we remove all Meta
		msgs := []openai.ChatCompletionMessage{}
		// if !ok {
		// 	logging.Info("coulnd not get []openai.ChatCompletionMessage")
		// }

		cancelCtx, cancel := context.WithCancel(ctx)

		msgs = append(msgs,
			openai.ChatCompletionMessage{
				Role:    roleToString(req.Role),
				Content: string(req.Content),
			})

		tools := []openai.Tool{}
		if req.F != nil {
			tools = append(tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        req.F.Name,
					Description: req.F.Parameters.Description,
					Parameters:  req.F.Parameters,
				},
			})
		}

		const ai_model = openai.GPT4oMini
		resp, err := p.clnt.CreateChatCompletion(cancelCtx,
			openai.ChatCompletionRequest{
				Model:    ai_model,
				Messages: msgs,
				Tools:    tools,
			},
		)

		if err != nil {
			logging.Info("creation with error", "err", err)

			return err
		}

		respMsg := resp.Choices[0].Message

		resp.Choices = resp.Choices[:1]

		if len(respMsg.ToolCalls) == 0 {
			return fmt.Errorf("Tool calls didn't called")
		}

		respMsg.ToolCalls = respMsg.ToolCalls[:1]

		msgs = append(msgs, respMsg)

		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    "",
			Name:       respMsg.ToolCalls[0].Function.Name,
			ToolCallID: respMsg.ToolCalls[0].ID,
		})

		select {
		case p.resps <- Resp{
			ReqData: req.ReqData,
			Content: model.Content(respMsg.Content),
			C:       model.Call(respMsg.ToolCalls[0].Function.Arguments),
			Err:     nil,
			Meta:    msgs}:

			logging.Info("put resp to thehe")

		case <-time.After(15 * time.Second):
			cancel()
			return fmt.Errorf("timeut")
		}

		return nil

	}

	if req.Tp == CheckSents {

		msgs, ok := req.Meta.([]openai.ChatCompletionMessage)
		if !ok {
			logging.Info("coulnd not get []openai.ChatCompletionMessage")
			return fmt.Errorf("coulnd not get []openai.ChatCompletionMessage")
		}

		cancelCtx, cancel := context.WithCancel(ctx)

		msgs = append(msgs,
			openai.ChatCompletionMessage{
				Role:    roleToString(req.Role),
				Content: string(req.Content),
			})

		tools := []openai.Tool{}
		if req.F != nil {
			tools = append(tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        req.F.Name,
					Description: req.F.Parameters.Description,
					Parameters:  req.F.Parameters,
				},
			})
		}

		const ai_model = openai.GPT4oMini
		resp, err := p.clnt.CreateChatCompletion(cancelCtx,
			openai.ChatCompletionRequest{
				Model:    ai_model,
				Messages: msgs,
				Tools:    tools,
			},
		)

		if err != nil {
			logging.Info("creation with error", "err", err)

			cancel()
			return err
		}

		respMsg := resp.Choices[0].Message

		if len(respMsg.ToolCalls) == 0 {
			cancel()
			return fmt.Errorf("Tool calls didn't called")
		}

		msgs = append(msgs, respMsg)

		select {
		case p.resps <- Resp{
			ReqData: req.ReqData,
			Content: model.Content(respMsg.Content),
			C:       model.Call(respMsg.ToolCalls[0].Function.Arguments),
			Err:     nil,
			Meta:    msgs}:

			logging.Info("put resp to thehe")

		case <-time.After(15 * time.Second):
			cancel()
			return fmt.Errorf("timeut")
		}

		return nil
	}

	return fmt.Errorf("unkown type")
}

func (g gateway) Run(ctx context.Context) {
	reqLstn := listener.New(
		g.reqs,
		processor.New(reqProcessor{resps: g.resps, clnt: g.clnt}, 16))

	reqLstn.Run(ctx)
}

type privateKey struct {
	Key string `json:"private_key"`
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
