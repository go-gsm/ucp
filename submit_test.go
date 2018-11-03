package ucp

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/go-gsm/charset"
)

func TestSubmit(t *testing.T) {
	actual := encodeMessage([]byte("01"), "Voyager", "09495696599", "Hello world",
		alphaNumericMessage, "", 23, 1, 1)
	data := struct {
		actual   []byte
		expected []byte
	}{
		actual,
		[]byte("\x0201/00126/O/51/09495696599/0ED6773E7C2ECB1B//1//1/////////////3/88/48656C6C6F20776F726C64////1////5039//020100060101070101///47\x03"),
	}

	if !bytes.Equal(data.expected, data.actual) {
		t.Errorf("Expected %s, got %s\n", data.expected, data.actual)
	}
}

func TestConvertToUCS2(t *testing.T) {
	ucs2TestCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"smiley emoji",
			"😃",
			"D83DDE03",
		},
		{
			`lower case letter "a"`,
			"a",
			"0061",
		},
	}

	for _, testCase := range ucs2TestCases {
		actual := fmt.Sprintf("%04X", string(charset.EncodeUcs2(testCase.input)))
		if actual != testCase.expected {
			t.Errorf("testcase %s: Expected %v, got %v\n", testCase.name, testCase.expected, actual)
		}
	}
}

func TestGetDataCodingScheme(t *testing.T) {
	dcsTestCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"alphanumeric",
			alphaNumericMessage,
			dataCodingSchemeASCII,
		},
		{
			"transparent data",
			transparentData,
			dataCodingSchemeUCS2,
		},
	}

	for _, testCase := range dcsTestCases {
		actual := getDataCodingScheme(testCase.input)
		if actual != testCase.expected {
			t.Errorf("testcase %s: Expected %v, got %v\n", testCase.name, testCase.expected, actual)
		}
	}
}

func TestGetMessageType(t *testing.T) {
	getMessageTypeTestCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"alpha numeric",
			"hello world",
			alphaNumericMessage,
		},
		{
			"unicode",
			"hello 世界",
			transparentData,
		},
	}

	for _, testCase := range getMessageTypeTestCases {
		actual := getMessageType(testCase.input)
		if actual != testCase.expected {
			t.Errorf("testcase %s: Expected %v, got %v\n", testCase.name, testCase.expected, actual)
		}
	}
}

func TestGetMessageParts(t *testing.T) {
	getMessagePartsTestCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			"ascii parts",
			"Did you ever hear the tragedy of Darth Plagueis The Wise? I thought not. It's not a story the Jedi would tell you. It's a Sith legend. Darth Plagueis was a Dark Lord of the Sith, so powerful and so wise he could use the Force to influence the midichlorians to create life... He had such a knowledge of the dark side that he could even keep the ones he cared about from dying. The dark side of the Force is a pathway to many abilities some consider to be unnatural. He became so powerful... the only thing he was afraid of was losing his power, which eventually, of course, he did. Unfortunately, he taught his apprentice everything he knew, then his apprentice killed him in his sleep. Ironic. He could save others from death, but not himself.",

			[]string{
				"Did you ever hear the tragedy of Darth Plagueis The Wise? I thought not. It's not a story the Jedi would tell you. It's a Sith legend. Darth Plagueis was",
				" a Dark Lord of the Sith, so powerful and so wise he could use the Force to influence the midichlorians to create life... He had such a knowledge of the ",
				"dark side that he could even keep the ones he cared about from dying. The dark side of the Force is a pathway to many abilities some consider to be unnat",
				"ural. He became so powerful... the only thing he was afraid of was losing his power, which eventually, of course, he did. Unfortunately, he taught his ap",
				"prentice everything he knew, then his apprentice killed him in his sleep. Ironic. He could save others from death, but not himself."},
		},

		{
			"unicode parts",
			"👌👀👌👀👌👀👌👀👌👀 good shit go౦ԁ sHit👌 thats ✔ some good👌👌shit right👌👌there👌👌👌 right✔there ✔✔if i do ƽaү so my self 💯 i say so 💯 thats what im talking about right there right there (chorus: ʳᶦᵍʰᵗ ᵗʰᵉʳᵉ) mMMMMᎷМ💯 👌👌 👌НO0ОଠOOOOOОଠଠOoooᵒᵒᵒᵒᵒᵒᵒᵒᵒ👌 👌👌 👌 💯 👌 👀 👀 👀 👌👌Good shit",
			[]string{
				"👌👀👌👀👌👀👌👀👌👀 good shit go౦ԁ sHit",
				"👌 thats ✔ some good👌👌shit right👌👌there👌👌",
				"👌 right✔there ✔✔if i do ƽaү so my self 💯 i say so ",
				"💯 thats what im talking about right there right there (chorus",
				": ʳᶦᵍʰᵗ ᵗʰᵉʳᵉ) mMMMMᎷМ💯 👌👌 👌НO0",
				"ОଠOOOOOОଠଠOoooᵒᵒᵒᵒᵒᵒᵒᵒᵒ👌 👌👌 ",
				"👌 💯 👌 👀 👀 👀 👌👌Good shit"},
		},
	}

	for _, testCase := range getMessagePartsTestCases {
		actual := getMessageParts(testCase.input)
		if len(testCase.expected) != len(actual) {
			t.Errorf("testcase %s: Expected %v, got %v\n", testCase.name, len(testCase.expected), len(actual))
		}
		for i := 0; i < len(testCase.expected); i++ {
			if testCase.expected[i] != actual[i] {
				t.Errorf("testcase %s: Expected %q, got %q\n", testCase.name, testCase.expected[i], actual[i])
			}
		}
	}
}

func BenchmarkEnodeMsg(b *testing.B) {
	message := "The quick brown fox jumps over the lazy dog is an English-language pangram - a sentence that contains all of the letters of the alphabet."
	sender := "Voyager"
	receiver := "09191234567"

	for n := 0; n < b.N; n++ {
		encodeMessage([]byte("01"), sender, receiver, message, alphaNumericMessage,
			"", 23, 1, 1)
	}
}
