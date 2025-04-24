package handler

import (
	"context"
	"fmt"

	"aidicti.top/api/protogen_cmn"
	"aidicti.top/api/protogen_gpt"
	"aidicti.top/gpt/internal/model"
	"aidicti.top/pkg/handler/proc"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
)

type controller interface {
	GetSentences(context.Context, pkgmodel.ReqData, *model.WordsContainer) (*model.SentenceContainer, error)
	CheckTranslations(context.Context, pkgmodel.ReqData, *model.TranslatedSentenceContainer) (*model.TranslationResultContainer, error)
}

type handlerProc_CheckTranslations interface {
	Exec(context.Context, *protogen_gpt.CheckTranslationsRequest) (*protogen_gpt.CheckTranslationsResponse, error)
}

type handlerProc_GetSentences interface {
	Exec(context.Context, *protogen_gpt.GetSentencesRequest) (*protogen_gpt.GetSentencesResponse, error)
}

type handler struct {
	protogen_gpt.UnimplementedAIProviderServiceServer
	ctrl controller

	procCheckTranslations handlerProc_CheckTranslations
	procGetSentences      handlerProc_GetSentences
}

func New(c controller) *handler {
	h := &handler{
		ctrl: c,
	}

	h.procCheckTranslations = proc.New(
		func(req *protogen_gpt.CheckTranslationsRequest) (*model.TranslatedSentenceContainer, error) {
			return &model.TranslatedSentenceContainer{Sentence: req.Sentences}, nil
		},

		func(resp *model.TranslationResultContainer) (*protogen_gpt.CheckTranslationsResponse, error) {
			out := &protogen_gpt.CheckTranslationsResponse{
				Id: &protogen_cmn.ReqData{},
			}

			utils.Contract(len(resp.Results) != 0)

			for _, res := range resp.Results {
				out.Sentences = append(out.Sentences, &protogen_gpt.SentenceCheck{
					Id:          0,
					Correction:  res.Correction,
					Explanation: res.Explaination,
					Rating:      uint64(res.Rating),
				})
			}

			return out, nil
		},

		h.ctrl.CheckTranslations,
	)

	h.procGetSentences = proc.New(
		func(req *protogen_gpt.GetSentencesRequest) (*model.WordsContainer, error) {
			utils.Contract(req != nil)

			if req.SentenceRequest == nil {
				return nil, fmt.Errorf("req.SentenceRequest != nil")
			}

			c := &model.WordsContainer{ID: model.ID(req.Id.ReqId)}

			for _, s := range req.SentenceRequest {
				word := model.Word{ID: model.ID(s.Id)}
				word.Word = s.Word
				word.Info = s.WordDescription
				c.Words = append(c.Words, word)
			}

			//NOTE do we really need to check it?
			utils.Contract(len(req.SentenceRequest) == len(c.Words))

			return c, nil
		},

		func(s *model.SentenceContainer) (*protogen_gpt.GetSentencesResponse, error) {
			utils.Contract(len(s.Sentences) != 0)

			sentences := []*protogen_gpt.Sentence{}

			for _, sentence := range s.Sentences {
				sentences = append(sentences, &protogen_gpt.Sentence{
					Id:          uint64(sentence.ID),
					Original:    sentence.Sentence,
					Word:        sentence.OriginalWord,
					Translation: sentence.Ideal})
			}

			r := protogen_gpt.GetSentencesResponse{Id: &protogen_cmn.ReqData{}, Sentences: sentences}

			//NOTE do we really need to check it?
			utils.Contract(len(s.Sentences) == len(sentences))

			return &r, nil
		},

		h.ctrl.GetSentences,
	)

	return h
}

func (h *handler) GetSentences(ctx context.Context,
	req *protogen_gpt.GetSentencesRequest) (*protogen_gpt.GetSentencesResponse, error) {

	return h.procGetSentences.Exec(ctx, req)
}

func (h *handler) CheckTranslations(
	ctx context.Context,
	req *protogen_gpt.CheckTranslationsRequest) (*protogen_gpt.CheckTranslationsResponse, error) {

	return h.procCheckTranslations.Exec(ctx, req)
}
