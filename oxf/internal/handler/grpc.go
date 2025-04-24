package handler

import (
	"context"
	"fmt"

	"aidicti.top/api/protogen_cmn"
	"aidicti.top/api/protogen_oxf"
	"aidicti.top/oxf/internal/model"
	"aidicti.top/pkg/handler/proc"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
)

type controller interface {
	GetDictEntry(ctx context.Context, ids pkgmodel.ReqData, word *model.Word) (
		*model.DictionaryEntry, error)
}

type handler struct {
	protogen_oxf.UnimplementedServiceOXFServer
	ctrl controller

	proc_GetDictEntry handlerProc_GetDictEntry
}

type handlerProc_GetDictEntry interface {
	Exec(context.Context, *protogen_oxf.Word) (*protogen_oxf.DictionaryEntry, error)
}

func New(c controller) *handler {
	h := &handler{
		ctrl: c,
	}

	h.proc_GetDictEntry = proc.New(
		func(req *protogen_oxf.Word) (*model.Word, error) {
			utils.Contract(req != nil)

			if req.Text == "" {
				return nil, fmt.Errorf("req.Text != \"\"")
			}

			return &model.Word{Text: req.Text}, nil
		},

		func(entry *model.DictionaryEntry) (*protogen_oxf.DictionaryEntry, error) {
			convertPronunciations := func(pronunciations []model.Pronunciation) []*protogen_oxf.Pronunciation {
				var protoPronunciations []*protogen_oxf.Pronunciation
				for _, p := range pronunciations {
					protoPronunciations = append(protoPronunciations, &protogen_oxf.Pronunciation{
						Lang:     p.Lang,
						Phonetic: p.Phonetic,
						Sound:    p.Sound,
					})
				}
				return protoPronunciations
			}

			convertExamples := func(examples []model.Example) []*protogen_oxf.Example {
				var protoExamples []*protogen_oxf.Example
				for _, e := range examples {
					protoExamples = append(protoExamples, &protogen_oxf.Example{
						Usage:   e.Usage,
						Example: e.Example,
					})
				}
				return protoExamples
			}

			convertSenses := func(senses []model.Sense) []*protogen_oxf.Sense {
				var protoSenses []*protogen_oxf.Sense
				for _, s := range senses {
					protoSenses = append(protoSenses, &protogen_oxf.Sense{
						Def:      s.Def,
						Examples: convertExamples(s.Examples),
						Pos:      int32(s.Pos),
					})
				}
				return protoSenses
			}

			convertSense := func(sense model.Sense) *protogen_oxf.Sense {
				return &protogen_oxf.Sense{
					Def:      sense.Def,
					Examples: convertExamples(sense.Examples),
					Pos:      int32(sense.Pos),
				}
			}

			convertIdioms := func(idioms []model.Idiom) []*protogen_oxf.Idiom {
				var protoIdioms []*protogen_oxf.Idiom
				for _, i := range idioms {
					protoIdioms = append(protoIdioms, &protogen_oxf.Idiom{
						Phrase: i.Phrase,
						Sense:  convertSense(i.Sense),
					})
				}
				return protoIdioms
			}

			convertRelatedWord := func(relatedWords []model.RelatedWord) []*protogen_oxf.RelatedWord {
				var protoRelatedWord []*protogen_oxf.RelatedWord
				for _, i := range relatedWords {
					protoRelatedWord = append(protoRelatedWord, &protogen_oxf.RelatedWord{
						Text: i.Text,
					})
				}
				return protoRelatedWord
			}

			protoEntry := &protogen_oxf.DictionaryEntry{
				Id:             &protogen_cmn.ReqData{},
				Word:           entry.Word,
				PartOfSpeech:   entry.PartOfSpeech,
				Pronunciations: convertPronunciations(entry.Pronunciation),
				Senses:         convertSenses(entry.Sences),
				Idioms:         convertIdioms(entry.Idioms),
				RelatedWords:   convertRelatedWord(entry.RelatedWords),
				Link:           entry.Link,
			}
			return protoEntry, nil
		},

		h.ctrl.GetDictEntry,
	)

	return h
}

func (h handler) GetDictEntry(
	ctx context.Context,
	word *protogen_oxf.Word) (*protogen_oxf.DictionaryEntry, error) {

	return h.proc_GetDictEntry.Exec(ctx, word)

}
