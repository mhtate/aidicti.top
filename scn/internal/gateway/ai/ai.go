package ai

import (
	"context"

	"aidicti.top/api/protogen_cmn"
	"aidicti.top/api/protogen_gpt"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/gatewaydialog/processor"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
	"aidicti.top/scn/internal/model"
)

type ReqAI struct {
	Id               pkgmodel.ReqData
	Words            *model.WordsReq
	SentencesToCheck string

	// AudioData model.AudioData
}

type RespAI struct {
	Id      pkgmodel.ReqData
	Sents   *model.SentsResp
	Results *model.TranslationResultContainer
}

type gateway struct {
	client    protogen_gpt.AIProviderServiceClient
	requests  chan ReqAI
	responses chan RespAI
}

func (g gateway) Reqs() chan<- ReqAI {
	return g.requests
}

func (g gateway) Resps() <-chan RespAI {
	return g.responses
}

func New(client protogen_gpt.AIProviderServiceClient) *gateway {
	return &gateway{
		client:    client,
		requests:  make(chan ReqAI, 64),
		responses: make(chan RespAI, 256),
	}
}

type reqProcessor struct {
	clnt  protogen_gpt.AIProviderServiceClient
	resps chan<- RespAI
}

// TODO we should use model here or there you kmowl
func (p reqProcessor) Process(ctx context.Context, req ReqAI) error {

	if req.Words != nil {
		utils.Assert(len(req.Words.Words) != 0, "container == 0")

		senReq := []*protogen_gpt.SentenceRequest{}

		for i, value := range req.Words.Words {
			senReq = append(senReq, &protogen_gpt.SentenceRequest{
				Id:              uint64(req.Id.ReqId) + uint64(i),
				Word:            value.Word,
				WordDescription: value.Info,
			})
		}

		//TODO request using like dialogId
		resp, err := p.clnt.GetSentences(ctx, &protogen_gpt.GetSentencesRequest{
			Id: &protogen_cmn.ReqData{
				ReqId: uint64(req.Id.ReqId),
				DlgId: uint64(req.Id.DlgId),
			},
			SentenceRequest: senReq,
		})

		if err != nil {
			logging.Info("ai request error", "err", err)
			return err
		}

		if resp.Sentences == nil {
			logging.Info("ai request error", "err", "resp.Sentences == nil")
			return err
		}

		if len(resp.Sentences) == 0 {
			logging.Info("ai request error", "err", "len(resp.Sentences) == 0")
			return err
		}

		sents := model.SentsResp{}

		for _, value := range resp.Sentences {
			sents.Sents = append(sents.Sents, model.SentResp{
				Original:    value.Original,
				Translation: value.Translation,
				Word:        value.Word,
			})
		}

		p.resps <- RespAI{
			Id: pkgmodel.ReqData{
				ReqId: pkgmodel.ReqID(resp.Id.ReqId),
				DlgId: pkgmodel.DlgID(resp.Id.DlgId),
			},
			Sents: &sents,
		}

		logging.Debug("ai received resp", "resp", resp)
	}

	if req.SentencesToCheck != "" {

		resp, err := p.clnt.CheckTranslations(ctx, &protogen_gpt.CheckTranslationsRequest{
			Id: &protogen_cmn.ReqData{
				ReqId: uint64(req.Id.ReqId),
				DlgId: uint64(req.Id.DlgId),
			},
			Sentences: req.SentencesToCheck,
		})

		if err != nil {
			logging.Info("ai request error", "err", err)
			return err
		}

		if resp.Sentences == nil {
			logging.Info("ai request error", "err", "resp.Sentences == nil")
			return err
		}

		if len(resp.Sentences) == 0 {
			logging.Info("ai request error", "err", "len(resp.Sentences) == 0")
			return err
		}

		result := model.TranslationResultContainer{}

		for _, value := range resp.Sentences {
			result.Results = append(result.Results, model.TranslationResult{
				Correction:   value.Correction,
				Explaination: value.Explanation,
				Rating:       uint8(value.Rating),
			})
		}

		p.resps <- RespAI{
			Id: pkgmodel.ReqData{
				ReqId: pkgmodel.ReqID(resp.Id.ReqId),
				DlgId: pkgmodel.DlgID(resp.Id.DlgId),
			},
			Results: &result,
		}

		logging.Debug("ai received resp", "resp", resp)
	}

	return nil
}

func (g gateway) Run(ctx context.Context) {
	reqLstn := listener.New(
		g.requests,
		processor.New(reqProcessor{clnt: g.client, resps: g.responses}, 16))

	reqLstn.Run(ctx)
}
