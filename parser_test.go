package ucp

import (
	"reflect"
	"testing"
)

func TestParseSessionResp(t *testing.T) {
	parseSessionRespTestCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"authenticated",
			"\x0200/00037/R/60/A/BIND AUTHENTICATED/6D\x03",
			"",
		},
		{
			"nack",
			"\x0200/00022/R/60/N/01/checksum error/04\x03",
			"[ucp error] code: 01 message: checksum error",
		},
		{
			"empty packet",
			"",
			"invalid packet",
		},
		{
			"invalid packet",
			"\x0204/00024/R/31/A/0003/5D\x03",
			"invalid packet",
		},
	}

	for _, testCase := range parseSessionRespTestCases {
		actual := parseSessionResp(testCase.input)
		if actual != nil && actual.Error() != testCase.expected {
			t.Errorf("testcase %s: Expected %v, got %v\n", testCase.name, testCase.expected, actual)
		}
	}
}

func TestParseResp(t *testing.T) {
	parseRespTestCases := []struct {
		name           string
		input          string
		expectedOpType string
		expectedFields []string
		expectedErr    error
	}{
		{
			"submit packet",
			"\x0202/00041/R/51/A//09495696599:120917113002/83\x03",
			opSubmitShortMessage,
			[]string{"02", "00041", "R", "51", "A", "", "09495696599:120917113002", "83"},
			nil,
		},
		{
			"empty packet",
			"",
			"",
			[]string{},
			errInvalidPacket,
		},
	}

	for _, testCase := range parseRespTestCases {
		opType, fields, err := parseResp(testCase.input)
		if opType != testCase.expectedOpType {
			t.Errorf("testcase %s: Expected %v, got %v\n", testCase.name, testCase.expectedOpType, opType)
		}
		if !reflect.DeepEqual(fields, testCase.expectedFields) {
			t.Errorf("testcase %s: Expected %v, got %v\n", testCase.name, testCase.expectedFields, fields)
		}
		if err != testCase.expectedErr {
			t.Errorf("testcase %s: Expected %v, got %v\n", testCase.name, testCase.expectedErr, err)
		}
	}
}
