package gateway

import "testing"

func TestCheckTranslations_Injection_Failure(t *testing.T) {

	gtw := New()

	const translationContent = "I translated the sentences to English. Check and correct all " +
		"types of mistakes I made. Explain every correction you made to help me in learning. " +
		"Rate my translation overall from 1-5 in integer numbers. " +
		"Return the response in the JSON format set_sentences(id, correction, explanation, rating). These my sentences: %s"

	gtw.NewDialog().Requests() <- AIRequest{}

}
