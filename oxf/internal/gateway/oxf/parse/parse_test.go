package parse_test

import (
	"os"
	"strings"
	"testing"

	"aidicti.top/oxf/internal/gateway/oxf/parse"
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestHTML(t *testing.T, filename string) *goquery.Document {
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
	require.NoError(t, err)

	return doc
}

func Test_GetEntry_Jittery_Adjective(t *testing.T) {
	doc := loadTestHTML(t, "testdata/oxf_jittery_a.html")
	html, _ := doc.Html()

	entry, err := parse.GetEntry(html)

	require.NoError(t, err)

	assert.Equal(t, "jittery", entry.Word)
	assert.Equal(t, "adjective", entry.PartOfSpeech)
	assert.Len(t, entry.Pronunciation, 2)
	assert.Len(t, entry.Sences, 1)
	assert.Len(t, entry.Sences[0].Examples, 4)
	assert.Len(t, entry.Idioms, 0)
}

func Test_GetEntry_Reflectance_Noun(t *testing.T) {
	doc := loadTestHTML(t, "testdata/oxf_reflectance_n.html")
	html, _ := doc.Html()

	entry, err := parse.GetEntry(html)
	require.NoError(t, err)

	assert.Equal(t, "reflectance", entry.Word)
	assert.Equal(t, "noun", entry.PartOfSpeech)
	assert.Len(t, entry.Pronunciation, 2)
	assert.Len(t, entry.Sences, 1)
	assert.Len(t, entry.Sences[0].Examples, 0)
	assert.Len(t, entry.Idioms, 0)
}

func Test_GetEntry_Get_Verb(t *testing.T) {
	doc := loadTestHTML(t, "testdata/oxf_get_v.html")
	html, _ := doc.Html()

	entry, err := parse.GetEntry(html)
	require.NoError(t, err)

	assert.Equal(t, "get", entry.Word)
	assert.Equal(t, "verb", entry.PartOfSpeech)
	assert.Len(t, entry.Pronunciation, 2)
	assert.Len(t, entry.Sences, 27)

	// to prepare a meal
	prepareMealSense := entry.Sences[19]
	assert.Len(t, prepareMealSense.Examples, 4)
	assert.Equal(t, "get something", prepareMealSense.Examples[0].Usage)
	assert.Equal(t, "", prepareMealSense.Examples[1].Usage)
	assert.Equal(t, "get something for somebody/yourself", prepareMealSense.Examples[2].Usage)
	assert.Equal(t, "get somebody/yourself something", prepareMealSense.Examples[3].Usage)
	assert.Equal(t, 20, prepareMealSense.Pos)

	// to reach a particular state or condition...
	reachParticularSense := entry.Sences[13]
	assert.Len(t, reachParticularSense.Examples, 17)
	assert.Equal(t, "", reachParticularSense.Grammar)
	assert.Equal(t, "", reachParticularSense.Examples[11].Usage)
	assert.Equal(t, 14, reachParticularSense.Pos)

	// to be connected with somebody by phone
	contactSense := entry.Sences[9]
	assert.Len(t, contactSense.Examples, 1)
	assert.Equal(t, "[transitive, no passive]", contactSense.Grammar)
	assert.Equal(t, "get somebody", contactSense.Usage)
	assert.Equal(t, "", contactSense.Examples[0].Usage)
	assert.Equal(t, 10, contactSense.Pos)
}

func Test_GetEntry_Table_Noun(t *testing.T) {
	doc := loadTestHTML(t, "testdata/oxf_table_n.html")
	html, _ := doc.Html()

	entry, err := parse.GetEntry(html)
	require.NoError(t, err)

	assert.Equal(t, "table", entry.Word)
	assert.Equal(t, "noun", entry.PartOfSpeech)
	assert.Len(t, entry.Pronunciation, 2)
	assert.Len(t, entry.Sences, 5)
	assert.Len(t, entry.Sences[1].Examples, 1)
	assert.Len(t, entry.Idioms, 9)
}

func Test_GetEntry_Bbb_Noun(t *testing.T) {
	doc := loadTestHTML(t, "testdata/oxf_bbb_n.html")
	html, _ := doc.Html()

	entry, err := parse.GetEntry(html)
	require.NoError(t, err)

	assert.Equal(t, "", entry.Word)
	assert.Equal(t, "", entry.PartOfSpeech)
	assert.Len(t, entry.RelatedWords, 10)
}
