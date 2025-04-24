package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"aidicti.top/gpt/internal/gateway"
	"aidicti.top/gpt/internal/gateway/gpt"
	"aidicti.top/gpt/internal/model"
	"aidicti.top/pkg/gatewaydialog"
	pkgdialog "aidicti.top/pkg/gatewaydialog/dialog"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/gatewaydialog/storage"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type AIGateway interface {
	NewDialog() gateway.Dialog
}

type gatewayGPT interface {
	Reqs() chan<- gpt.Req
	Resps() <-chan gpt.Resp
}

type storageGPT = gatewaydialog.Storage[gpt.Req, gpt.Resp]

//TODO explain a new word with your own words it is a good idea

type definition struct {
	Type        string                `json:"type"`
	Properties  map[string]definition `json:"properties,omitempty"`
	Items       *definition           `json:"items,omitempty"`
	Required    []string              `json:"required,omitempty"`
	Description string                `json:"description,omitempty"`
}

// Define a struct to represent a sentence object in the JSON
type sentence struct {
	ID           int    `json:"id"`
	OriginalWord string `json:"originalWord"`
	Ideal        string `json:"idealTranslationInEnglish"`
	Sentence     string `json:"sentence"`
}

// Define a struct to represent the root object in the JSON
type root struct {
	Sentences []sentence `json:"sentences"`
}

// Sentence represents a single sentence with its details
type result struct {
	Correction  string `json:"correction"`
	Explanation string `json:"explanation"`
	Rating      uint8  `json:"rating"`
}

// TODO rename please it
// SentencesWrapper represents the top-level object containing a list of sentences
type rootTr struct {
	Results []result `json:"sentences"`
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

func toMessage(c model.WordsContainer) (string, error) {
	output := ""

	//TODO haha do it normal
	for _, word := range c.Words {
		output += word.Word + "(" + word.Info + "), "
	}

	return output, nil
}

func toMessageT(c model.TranslatedSentenceContainer) (string, error) {
	return c.Sentence, nil
}

type service struct {
	channels map[model.ID](gateway.Dialog)
	gateway  AIGateway

	gtwGPT  gatewayGPT
	strgGPT storageGPT
}

type respProcGPT struct {
	strg storageGPT
}

type reqCnvGPT struct {
	id pkgmodel.DlgID
	d  *dialogGPT
}

func (c reqCnvGPT) Convert(req gpt.Req) (gpt.Req, error) {
	reqId := pkgmodel.ReqID(rand.Uint64())

	req.ReqData = pkgmodel.ReqData{
		DlgId: c.id,
		ReqId: reqId}

	req.Meta = c.d.Meta

	return req, nil
}

func (c *respProcGPT) Process(ctx context.Context, resp gpt.Resp) error {
	dialog, err := c.strg.Get(resp.DlgId)
	if err != nil {
		logging.Warn("get dlg from stg", "st", "fail", "dlgId", resp.DlgId, "reqId", resp.ReqId, "err", err)
		return err
	}

	//NOTE crutch
	{
		d, ok := dialog.(*dialogGPT)
		if !ok {
			return fmt.Errorf("to dialog gpt fail")
		}

		// resp.Meta = d.Meta
		d.Meta = resp.Meta
	}

	select {
	case dialog.Resps() <- resp:
		logging.Debug("pass resp to dlg", "st", "ok", "dlgId", resp.DlgId)

	default:
		logging.Warn("pass resp to dlg", "st", "fail", "rsn", "ch closed or full")
		return fmt.Errorf("asdasdasdasdasd")
	}

	return nil
}

type dialogGPT struct {
	dialog gatewaydialog.Dialog[gpt.Req, gpt.Resp]
	//NOTE still not sure
	Meta any
}

func (d dialogGPT) Run(ctx context.Context) {
	d.dialog.Run(ctx)
}

func (d dialogGPT) Reqs() chan gpt.Req {
	return d.dialog.Reqs()
}

func (d dialogGPT) Resps() chan gpt.Resp {
	return d.dialog.Resps()
}

func NewService(g AIGateway, gpt_ gatewayGPT) service {
	return service{
		channels: make(map[model.ID]gateway.Dialog),
		gateway:  g,
		gtwGPT:   gpt_,
		strgGPT: storage.New(context.TODO(),
			func(id pkgmodel.DlgID) (
				gatewaydialog.Dialog[gpt.Req, gpt.Resp], error) {

				d := dialogGPT{}

				cnv_ := reqCnvGPT{id: id, d: &d}

				d.dialog = pkgdialog.New[gpt.Resp](id, gpt_, cnv_)

				logging.Debug("create dlg", "st", "ok", "id", id)

				return &d, nil
			}),
	}
}

// TODO move it to database or some storage
// TODO try ask for optional words in sentences that may remeber something to learn
// TODO ask about alternative versions of the same sentence
// const completionContent = "Create sentences in %s that include translations of the following " +
// 	"English words: %s.\nReturn the response in the JSON format set_sentences(id, original_word_required, english_sentence, russian_sentence)"

// const completionContent = "Create several sentences on %s language, " +
// 	"that contain translations of these English words (if there are parentheses after " +
// 	"a word or expression, this means that you need to use exactly this meaning of this word " +
// 	"or/and part of speech): %s. Return the answer in json format sentences: {\"originalWord\": ?, \"sentence\": ?, \"idealTranslationInEnglish\": ?, }"

const completionContent = "Create sentence on %s language, " +
	"that contain translations of these English words (if there are parentheses after " +
	"a word or expression, this means that you need to use exactly this meaning of this word " +
	"or/and part of speech): %s. Return the answer in json format sentences: {\"originalWord\": ?, \"sentence\": ?, \"idealTranslationInEnglish\": ?, }"

const translationContent = "I translated the sentences to English. Check and correct all " +
	"types of mistakes I made. Explain every correction you made to help me in learning. " +
	"Rate my translation overall from 1-5 in integer numbers. " +
	"Return the response in the JSON format sentences: {\"correction\": ?, \"explanation\": ?, \"rating\": ?}. These my sentences: %s"

func (s service) GetSentences(ctx context.Context, id_ pkgmodel.ReqData, c model.WordsContainer) (model.SentenceContainer, error) {
	// dialog := s.gateway.NewDialog()
	// s.channels[c.ID] = dialog

	cancelCtx, _ := context.WithCancel(ctx)

	const originalWord = "originalWord"
	const id = "id"
	const sentence = "idealTranslationInEnglish"
	const russian_sentence = "sentence"
	const sentences = "sentences"

	getJsonFuncSchema := func() *jsonschema.Definition {
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
						Required: []string{originalWord, sentence, russian_sentence},
					},
				},
			},
			Required: []string{sentences},
		}

		return &schema
	}

	dlg, err := s.strgGPT.Get(pkgmodel.DlgID(id_.DlgId))
	if err != nil {
		fmt.Errorf("can't create dialog")
	}

	//TODO error check
	strings, _ := toMessage(c)

	dlg.Reqs() <- gpt.Req{
		Context: cancelCtx,
		Content: model.Content(fmt.Sprintf(completionContent, "Russian", strings)),
		F:       &model.Func{Name: "sentences", Description: "sentences: {\"originalWord\": ?, \"sentence\": ?, \"idealTranslationInEnglish\": ?, }", Parameters: *getJsonFuncSchema()},
		Role:    model.RoleAssistant,
		Tp:      gpt.GetSents,
	}

	select {
	case resp := <-dlg.Resps():
		if resp.Err != nil {
			return model.SentenceContainer{}, fmt.Errorf("task is finished with an error")
		}

		var root root
		err := json.Unmarshal([]byte(resp.C), &root)
		if err != nil {
			return model.SentenceContainer{}, fmt.Errorf("task is finished with an error")
		}

		//TODO put value inside
		c := model.SentenceContainer{ID: c.ID, Sentences: make([]model.Sentence, 0, len(root.Sentences))}
		for _, sentence := range root.Sentences {
			//TODO do we really need it (id)
			c.Sentences = append(c.Sentences, model.Sentence{
				ID:           c.ID,
				Sentence:     sentence.Sentence,
				OriginalWord: sentence.OriginalWord,
				Ideal:        sentence.Ideal,
			})
		}

		return c, nil

	case <-ctx.Done():
		return model.SentenceContainer{}, fmt.Errorf("task is cancelled")
	case <-time.After(10 * time.Second):
		return model.SentenceContainer{}, fmt.Errorf("task is cancelled")
	}

	return model.SentenceContainer{}, nil
}

func (s service) CheckTranslations(ctx context.Context, reqData pkgmodel.ReqData, c model.TranslatedSentenceContainer) (model.TranslationResultContainer, error) {
	// dialog, ok := s.channels[c.ID]
	// if !ok {
	// 	return model.TranslationResultContainer{}, fmt.Errorf("dialog is not found")
	// }

	cancelCtx, _ := context.WithCancel(ctx)

	const id = "id"
	const correction = "correction"
	const explanation = "explanation"
	const rating = "rating"
	const sentences = "sentences"

	getJsonFuncSchema := func() *jsonschema.Definition {
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
							correction: {
								Type:        jsonschema.String,
								Description: correction,
							},
							explanation: {
								Type:        jsonschema.String,
								Description: explanation,
							},
							rating: {
								Type:        jsonschema.Integer,
								Description: rating,
							},
						},
						Required: []string{correction, explanation, rating},
					},
				},
			},
			Required: []string{sentences},
		}

		return &schema
	}

	dlg, err := s.strgGPT.Get(pkgmodel.DlgID(reqData.DlgId))
	if err != nil {
		fmt.Errorf("can't create dialog")
	}

	// strings, _ := toMessageT(c)

	dlg.Reqs() <- gpt.Req{
		Context: cancelCtx,
		Content: model.Content(fmt.Sprintf(translationContent, c)),
		F:       &model.Func{Name: "sentences", Description: "sentences: {\"correction\": ?, \"explanation\": ?, \"rating\": ?}", Parameters: *getJsonFuncSchema()},
		Role:    model.RoleAssistant,
		Tp:      gpt.CheckSents,
	}

	select {
	case response := <-dlg.Resps():
		if response.Err != nil {
			return model.TranslationResultContainer{}, fmt.Errorf("task is finished with an error")
		}

		var root rootTr
		err := json.Unmarshal([]byte(response.C), &root)
		if err != nil {
			return model.TranslationResultContainer{}, fmt.Errorf("task is finished with an error")
		}

		c := model.TranslationResultContainer{Results: make([]model.TranslationResult, 0, len(root.Results))}
		for _, result := range root.Results {
			c.Results = append(c.Results, model.TranslationResult{Explaination: result.Explanation, Correction: result.Correction, Rating: result.Rating})
		}

		return c, nil

	case <-ctx.Done():
		return model.TranslationResultContainer{}, fmt.Errorf("task is cancelled")
	case <-time.After(15 * time.Second):
		return model.TranslationResultContainer{}, fmt.Errorf("task is cancelled")
	}

	return model.TranslationResultContainer{}, nil

}

func (s service) Run(ctx context.Context) {
	procRespGPT := respProcGPT{s.strgGPT}
	lstnGPT := listener.New(s.gtwGPT.Resps(), &procRespGPT)
	go lstnGPT.Run(ctx)

	select {
	case <-ctx.Done():
	}
}
