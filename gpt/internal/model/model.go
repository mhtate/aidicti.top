package model

import (
	"fmt"

	"github.com/sashabaranov/go-openai/jsonschema"
)

type ID uint64

type Word struct {
	ID   ID
	Word string
	Info string
}

type WordsContainer struct {
	ID    ID
	Words []Word
}

type Sentence struct {
	ID           ID
	Sentence     string
	OriginalWord string
	Ideal        string
}

type SentenceContainer struct {
	ID        ID
	Sentences []Sentence
}

func (c *SentenceContainer) String() string {

	sntcString := ""
	for _, sntc := range c.Sentences {
		sntcString += fmt.Sprintf(
			"\tid: %d,\n\toriginal: %s,\n\tsentence: %s\n",
			sntc.ID, sntc.OriginalWord, sntc.Sentence)
	}

	return fmt.Sprintf("id: %d,\n%s", c.ID, sntcString)
}

type TranslatedSentence struct {
	ID          ID
	ID_original ID
	Sentence    string
}

type TranslatedSentenceContainer struct {
	Sentence string
}

type TranslationResult struct {
	Correction   string
	Explaination string
	Rating       uint8
}

type TranslationResultContainer struct {
	ID          ID
	ID_original ID
	Results     []TranslationResult
}

type Role int

const (
	RoleSystem Role = iota
	RoleUser
	RoleAssistant
	RoleFunction
	RoleTool
	RoleUnknown
)

type Function struct {
	ID ID
}

type Message struct {
	Role         Role
	Content      string
	Functions    []Function
	FinishReason string
}

type Dialog struct {
	ID       ID
	Messages []string
	Role     Role
}

type Call string

type Content string

type Func struct {
	Name        string
	Description string
	Parameters  jsonschema.Definition
}
