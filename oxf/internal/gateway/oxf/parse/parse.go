package parse

import (
	"fmt"
	"strconv"
	"strings"

	"aidicti.top/oxf/internal/model"
	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

func getIdioms(sel *goquery.Selection) []model.Idiom {
	if idiomsS := sel.Find("div.idioms").First(); idiomsS != nil {
		return getidioms(idiomsS)
	}

	return []model.Idiom{}
}

func getidioms(sel *goquery.Selection) []model.Idiom {
	idioms := []model.Idiom{}
	sel.Find("span.idm-g").Each(func(i int, s *goquery.Selection) {

		idiom := model.Idiom{}

		idiom.Phrase = s.Find("span.idm").First().Text()
		// <span class="labels" htag="span" hclass="labels">(informal)</span>

		idiom.Sense = getSense(s, -1)

		idioms = append(idioms, idiom)
	})

	return idioms
}

func getSense(sel *goquery.Selection, pos int) model.Sense {
	utils.Assert(sel != nil, "get nil pointer *goquery.Selection")

	sense := model.Sense{Examples: []model.Example{}, Pos: pos}

	//definition
	defNode := sel.Find("span.def").First()

	// <span class="grammar" htag="span" hclass="grammar">[transitive, no passive]</span>
	grammarNode := sel.Find("span.grammar").First()

	// <span class="cf" hclass="cf" htag="span">get somebody</span>
	cfNode := sel.Find("span.cf").First()

	sense.Def = defNode.Text()
	sense.Grammar = grammarNode.Text()
	if (cfNode.Parent() != nil) && (cfNode.Parent().Nodes != nil) &&
		(cfNode.Parent().Nodes[0] == defNode.Parent().Nodes[0]) {
		sense.Usage = cfNode.Text()
	}

	//examples with using
	parent := sel.Find("ul.examples").First()

	//extra examples
	extraExamples := sel.Find("ul.examples").Eq(1)

	parent.Find("li").Each(func(i int, s *goquery.Selection) {
		example := model.Example{}

		example.Usage = s.Find("span.cf").First().Text()

		example.Example = s.Find("span.x").First().Text()

		sense.Examples = append(sense.Examples, example)

	})

	extraExamples.Find("li").Each(func(i int, s *goquery.Selection) {
		example := model.Example{}

		example.Example = s.Find("span.unx").First().Text()

		sense.Examples = append(sense.Examples, example)

	})

	return sense
}

func getSensens(sel *goquery.Selection) []model.Sense {
	senses := []model.Sense{}

	firstS := sel.Find("li.sense").First()
	_, exists := firstS.Attr("sensenum")
	if !exists {
		return []model.Sense{getSense(firstS, 0)}
	}

	sensesS := sel.Find("li.sense").FilterFunction(
		func(i int, s *goquery.Selection) bool {
			_, exists := s.Attr("sensenum")
			return exists
		})

	const NotSet = -1
	prevNum := NotSet
	sensesS.Each(func(i int, s *goquery.Selection) {
		val, exists := s.Attr("sensenum")
		utils.Assert(exists, "")

		i, err := strconv.Atoi(val)
		if err != nil {
			return
		}

		if (prevNum == NotSet) || (i == prevNum+1) {
			prevNum = i
		} else {
			return
		}

		senses = append(senses, getSense(s, i))
	})

	return senses
}

func getPronunciation(sel *goquery.Selection) []model.Pronunciation {
	prons := []model.Pronunciation{}

	ukNode := sel.Find(".phons_br").First()
	if ukNode != nil {

		sound, _ := ukNode.Find("div.sound").First().Attr("data-src-mp3")
		phonetic := ukNode.Find("span.phon").First().Text()

		prons = append(prons, model.Pronunciation{
			Lang:     "en_GB",
			Sound:    sound,
			Phonetic: phonetic})
	}

	usNode := sel.Find(".phons_n_am").First()
	if usNode != nil {

		sound, _ := ukNode.Find("div.sound").First().Attr("data-src-mp3")
		phonetic := ukNode.Find("span.phon").First().Text()

		prons = append(prons, model.Pronunciation{
			Lang:     "en_US",
			Sound:    sound,
			Phonetic: phonetic})
	}

	return prons
}

func GetEntry(html string) (*model.DictionaryEntry, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logging.Debug("parse html fail", "err", err)
		return nil, err
	}

	wordNode := doc.Find("h1.headword")
	if wordNode.Length() == 0 {
		didyoumeanNode := doc.Find("#didyoumean")
		if didyoumeanNode.Length() == 0 {
			return nil, fmt.Errorf("didn't find anything")
		}

		relatedWord := []model.RelatedWord{}

		// we didn't find the word but the site answers with similar words
		doc.Find("a.dym-link").Each(func(i int, s *goquery.Selection) {
			relatedWord = append(relatedWord, model.RelatedWord{Text: s.Text()})
		})

		//TODO i think we should send an error here
		return &model.DictionaryEntry{
			RelatedWords: relatedWord,
		}, nil
	}

	word := wordNode.Text()

	// Verb Get sense 14, there is additional cursive (linking verb) with the same
	//  tags and classes as main part of speech, that why we pick only first one
	// 	- <span class="pos" htag="span" hclass="pos">linking verb</span>
	// 	- <span class="pos" htag="span" hclass="pos">verb</span>
	partOfSpeech := doc.Find("span.pos").First().Text()

	// verbForms := make(map[string]string)
	// doc.Find(".verb_forms_table tr").Each(func(i int, s *goquery.Selection) {
	// 	form := s.Find("td .vf_prefix").Text()
	// 	verb := s.Find("td .verb_form").Text()
	// 	verbForms[form] = strings.TrimSpace(verb)
	// })

	return &model.DictionaryEntry{
		Word:          word,
		PartOfSpeech:  partOfSpeech,
		Pronunciation: getPronunciation(doc.Selection),
		// VerbForms:     verbForms,
		Sences:       getSensens(doc.Selection),
		Idioms:       getIdioms(doc.Selection),
		RelatedWords: []model.RelatedWord{},
	}, nil
}
