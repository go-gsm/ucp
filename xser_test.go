package ucp

import (
	"reflect"
	"testing"
)

func TestBuildXserUDH(t *testing.T) {
	xserUdhTestCases := []struct {
		refNum        int
		totalMsgParts int
		msgPartNum    int
		expected      string
	}{
		{
			42,
			1,
			1,
			"",
		},
		{
			42,
			2,
			1,
			"01060500032A0201",
		},
	}

	for _, testCase := range xserUdhTestCases {
		actual := buildXSerUDH(testCase.refNum, testCase.totalMsgParts, testCase.msgPartNum)
		if actual != testCase.expected {
			t.Errorf("Expected %s, got %s\n", testCase.expected, actual)
		}
	}
}

func TestBuildXserBillingID(t *testing.T) {

	billingIDTests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"01000001C123000210", "0C12303130303030303143313233303030323130"},
		{"01000001C1230001F0", "0C12303130303030303143313233303030314630"},
	}

	for _, testCase := range billingIDTests {
		actual := buildXSERBillingID(testCase.input)
		if actual != testCase.expected {
			t.Errorf("Expected %s, got %s\n", testCase.expected, actual)
		}
	}
}

func TestParseXser(t *testing.T) {
	parseXserTestCases := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			"empty extra service data",
			"",
			map[string]string{},
		},
		{
			"gsm data coding scheme info",
			"020100",
			map[string]string{
				"02": "00",
			},
		},
	}

	for _, testCase := range parseXserTestCases {
		actual := parseXser(testCase.input)
		if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("Expected %v, got %v\n", testCase.expected, actual)
		}
	}
}
