package controller

import (
	"context"

	"aidicti.top/gpt/internal/model"
	pkgmodel "aidicti.top/pkg/model"
)

type service interface {
	GetSentences(context.Context, pkgmodel.ReqData, model.WordsContainer) (model.SentenceContainer, error)
	CheckTranslations(context.Context, pkgmodel.ReqData, model.TranslatedSentenceContainer) (model.TranslationResultContainer, error)
}

type dialogStorage interface {
	Create() (model.Dialog, error)
	Get(model.ID) (model.Dialog, error)
	Delete(model.Dialog) error
	Store(model.Dialog) error
}

type controller struct {
	gateway service
	storage dialogStorage
}

func toMessage(c model.WordsContainer) (string, error) {
	output := ""

	//TODO haha do it normal
	for _, word := range c.Words {
		output += word.Word + "(" + word.Info + "), "
	}

	return output, nil
}

func New(g service, s dialogStorage) *controller {
	return &controller{g, s}
}

const completionContent = "Create sentences in %s that include translations of the following " +
	"English words: %s.\nReturn the response in the JSON format, using openai Tools"

func (c *controller) GetSentences(ctx context.Context, id pkgmodel.ReqData, words *model.WordsContainer) (*model.SentenceContainer, error) {

	if len(words.Words) == 0 {
		panic("Violated! len(words.Words) == 0")
	}

	//TODO it's service logic & put language pick

	// strings, _ := toMessage(words)

	// out := fmt.Sprintf(completionContent, "Russian", strings)

	// dialog.Messages = append(dialog.Messages, out)

	sentences, err := c.gateway.GetSentences(ctx, id, *words)

	if err != nil {
		return &model.SentenceContainer{}, err
	}

	return &sentences, nil

	// if err != nil {
	// 	//TODO types of error

	// 	return model.SentenceContainer{}, err
	// }

	// if len(sentences.Sentences) == 0 {
	// 	//TODO types of error
	// 	return model.SentenceContainer{}, err
	// }

	// err = c.storage.Store(dialog)
	// if err != nil {CheckResults
	// 	//TODO types of error

	// 	return model.SentenceContainer{}, err
	// }

	// return sentences, nil
}

func (c *controller) CheckTranslations(ctx context.Context, reqData pkgmodel.ReqData, translations *model.TranslatedSentenceContainer) (*model.TranslationResultContainer, error) {
	// if len(translations.Sentences) == 0 {
	// 	panic("Violated! len(translations.Sentences) == 0")
	// }

	// _, err := c.storage.Get(translations.ID_original)

	// if err != nil {
	// 	//TODO types of error
	// 	return model.TranslationResultContainer{}, err
	// }

	// checks, err := c.gateway.CheckTranslations(ctx, translations)

	// if err != nil {
	// 	//TODO types of error
	// 	return model.CheckResults{}, err
	// }

	// err = c.storage.Delete(dialog)
	// if err != nil {
	// 	//TODO types of error
	// 	return model.CheckResults{}, err
	// }

	val, err := c.gateway.CheckTranslations(ctx, reqData, *translations)

	return &val, err
}
