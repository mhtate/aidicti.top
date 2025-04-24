package conv

import (
	"aidicti.top/api/protogen_oxf"
	"aidicti.top/pkg/model/oxf"
)

func FromProto(protoEntry *protogen_oxf.DictionaryEntry) *oxf.DictionaryEntry {
	if protoEntry == nil {
		return nil
	}

	modelEntry := &oxf.DictionaryEntry{
		Word:         protoEntry.Word,
		PartOfSpeech: protoEntry.PartOfSpeech,
		Link:         protoEntry.Link,
	}

	for _, protoPronunciation := range protoEntry.Pronunciations {
		modelEntry.Pronunciations = append(modelEntry.Pronunciations, oxf.Pronunciation{
			Lang:     protoPronunciation.Lang,
			Phonetic: protoPronunciation.Phonetic,
			Sound:    protoPronunciation.Sound,
		})
	}

	for _, protoSense := range protoEntry.Senses {
		modelSense := oxf.Sense{
			Def:     protoSense.Def,
			Usage:   protoSense.Usage,
			Grammar: protoSense.Grammar,
			Pos:     int(protoSense.Pos),
		}

		for _, protoExample := range protoSense.Examples {
			modelSense.Examples = append(modelSense.Examples, oxf.Example{
				Usage:   protoExample.Usage,
				Example: protoExample.Example,
			})
		}

		modelEntry.Senses = append(modelEntry.Senses, modelSense)
	}

	for _, protoIdiom := range protoEntry.Idioms {
		modelEntry.Idioms = append(modelEntry.Idioms, oxf.Idiom{
			Phrase: protoIdiom.Phrase,
			Sense: oxf.Sense{
				Def:     protoIdiom.Sense.Def,
				Usage:   protoIdiom.Sense.Usage,
				Grammar: protoIdiom.Sense.Grammar,
				Pos:     int(protoIdiom.Sense.Pos),
			},
		})
	}

	for _, protoRelatedWord := range protoEntry.RelatedWords {
		modelEntry.RelatedWords = append(modelEntry.RelatedWords, oxf.RelatedWord{
			Text: protoRelatedWord.Text})
	}

	return modelEntry
}

func ToProto(modelEntry *protogen_oxf.DictionaryEntry) *oxf.DictionaryEntry {
	if modelEntry == nil {
		return nil
	}

	protoEntry := &oxf.DictionaryEntry{
		Word:         modelEntry.Word,
		PartOfSpeech: modelEntry.PartOfSpeech,
		Link:         modelEntry.Link,
	}

	for _, modelPronunciation := range modelEntry.Pronunciations {
		protoEntry.Pronunciations = append(protoEntry.Pronunciations, oxf.Pronunciation{
			Lang:     modelPronunciation.Lang,
			Phonetic: modelPronunciation.Phonetic,
			Sound:    modelPronunciation.Sound,
		})
	}

	for _, modelSense := range modelEntry.Senses {
		protoSense := oxf.Sense{
			Def:     modelSense.Def,
			Usage:   modelSense.Usage,
			Grammar: modelSense.Grammar,
			Pos:     int(modelSense.Pos),
		}
		for _, modelExample := range modelSense.Examples {
			protoSense.Examples = append(protoSense.Examples, oxf.Example{
				Usage:   modelExample.Usage,
				Example: modelExample.Example,
			})
		}
		protoEntry.Senses = append(protoEntry.Senses, protoSense)
	}

	for _, modelIdiom := range modelEntry.Idioms {
		protoEntry.Idioms = append(protoEntry.Idioms, oxf.Idiom{
			Phrase: modelIdiom.Phrase,
			Sense: oxf.Sense{
				Def:     modelIdiom.Sense.Def,
				Usage:   modelIdiom.Sense.Usage,
				Grammar: modelIdiom.Sense.Grammar,
				Pos:     int(modelIdiom.Sense.Pos),
			},
		})
	}

	for _, modelRelatedWord := range modelEntry.RelatedWords {
		protoEntry.RelatedWords = append(protoEntry.RelatedWords, oxf.RelatedWord{
			Text: modelRelatedWord.Text})
	}

	return protoEntry
}
