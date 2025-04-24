package oxf

type Example struct {
	Usage   string
	Example string
}

type Sense struct {
	Def      string
	Usage    string
	Grammar  string
	Examples []Example
	Pos      int
}

type Idiom struct {
	Phrase string
	Sense  Sense
}

type Pronunciation struct {
	Lang     string
	Phonetic string
	Sound    string
}

type DictionaryEntry struct {
	Word           string
	PartOfSpeech   string
	Pronunciations []Pronunciation
	// VerbForms     map[string]string
	Senses       []Sense
	Idioms       []Idiom
	RelatedWords []RelatedWord
	Link         string
}

type Word struct {
	Text string
}

type RelatedWord struct {
	Text string
}
