package ucp

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

func TestDeliverSm(t *testing.T) {
	buf := new(bytes.Buffer)
	writer := bufio.NewWriter(buf)
	wg := new(sync.WaitGroup)
	closeChan := make(chan struct{}, 1)
	deliverMsgCh := make(chan []string, 1)
	deliverMsgPartCh := make(chan deliverMsgPart, 1)
	deliverMsgCompleteCh := make(chan deliverMsgPart, 1)
	client := &Client{
		muconn: &sync.Mutex{},
		logger: log.New(os.Stdout, "debug ", 0),
	}
	readDeliveryMsg(writer, wg, closeChan, deliverMsgCh, deliverMsgPartCh, deliverMsgCompleteCh,
		client.muconn, client)
	runtime.Gosched()

	mobileOriginatingMessage := []string{"26", "00408", "O", "52", "2371", "09191234567", "", "", "", "", "", "", "", "", "", "", "", "0000", "121017010208", "", "", "", "3", "", "41414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141", "", "", "0", "", "", "", "", "", "", "020100", "", "", "BF"}

	deliverMsgCh <- mobileOriginatingMessage
	actual := <-deliverMsgCompleteCh
	expected := deliverMsgPart{
		currentPart: 0,
		totalParts:  0,
		sender:      "09191234567",
		receiver:    "2371",
		message:     "41414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141", msgID: "09191234567:121017010208",
		dcs: "00",
	}

	close(closeChan)
	wg.Wait()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %q, got %q\n", expected, actual)
	}

	expectedBytesWritten := []byte("\x0226/00037/R/52/A//2371:121017010208/03\x03")
	actualBytesWritten := buf.Bytes()
	if !bytes.Equal(expectedBytesWritten, actualBytesWritten) {
		t.Errorf("Expected %v, got %v\n", expectedBytesWritten, actualBytesWritten)
	}
}

func TestDeliverSmMultiPartIncomplete(t *testing.T) {

	buf := new(bytes.Buffer)
	writer := bufio.NewWriter(buf)
	wg := new(sync.WaitGroup)
	closeChan := make(chan struct{}, 1)
	deliverMsgCh := make(chan []string, 1)
	deliverMsgPartCh := make(chan deliverMsgPart, 1)
	deliverMsgCompleteCh := make(chan deliverMsgPart, 1)

	client := &Client{
		muconn: &sync.Mutex{},
		logger: log.New(os.Stdout, "debug ", 0),
	}
	readDeliveryMsg(writer, wg, closeChan, deliverMsgCh, deliverMsgPartCh, deliverMsgCompleteCh,
		client.muconn, client)
	runtime.Gosched()

	mobileOriginatingMessage := []string{"05", "00410", "O", "52", "2371", "09191234567", "", "", "", "", "", "", "", "", "", "", "", "0000", "290917182523", "", "", "", "3", "", "44696420796F7520657665722068656172207468652074726167656479206F6620446172746820506C6167756569732054686520576973653F20492074686F75676874206E6F742E2049742773206E6F7420612073746F727920746865204A65646920776F756C642074656C6C20796F752E204974277320612053697468206C6567656E642E20446172746820506C61677565697320776173", "", "", "0", "", "", "", "", "", "", "01060500036D0501020100", "", "", "21"}

	deliverMsgCh <- mobileOriginatingMessage
	actual := <-deliverMsgPartCh

	expected := deliverMsgPart{
		currentPart: 1,
		totalParts:  5,
		refNum:      0x6d,
		sender:      "09191234567",
		receiver:    "2371",
		message:     "44696420796F7520657665722068656172207468652074726167656479206F6620446172746820506C6167756569732054686520576973653F20492074686F75676874206E6F742E2049742773206E6F7420612073746F727920746865204A65646920776F756C642074656C6C20796F752E204974277320612053697468206C6567656E642E20446172746820506C61677565697320776173", msgID: "09191234567:290917182523",
		dcs: "00",
	}

	close(closeChan)
	wg.Wait()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %q, got %q\n", expected, actual)
	}

	expectedBytesWritten := []byte("\x0205/00037/R/52/A//2371:290917182523/1A\x03")
	actualBytesWritten := buf.Bytes()
	if !bytes.Equal(expectedBytesWritten, actualBytesWritten) {
		t.Errorf("Expected %v, got %v\n", expectedBytesWritten, actualBytesWritten)
	}
}

func TestDeliverSmMultiPartComplete(t *testing.T) {

	wg := new(sync.WaitGroup)
	closeChan := make(chan struct{}, 1)

	deliverMsgPartCh := make(chan deliverMsgPart, 1)
	deliverMsgCompleteCh := make(chan deliverMsgPart, 1)
	readPartialDeliveryMsg(wg, closeChan, deliverMsgPartCh, deliverMsgCompleteCh,
		&Client{logger: log.New(os.Stdout, "debug ", 0)})
	runtime.Gosched()

	deliverMsgPartCh <- deliverMsgPart{
		currentPart: 1,
		totalParts:  3,
		refNum:      109,
		sender:      "09191234567",
		receiver:    "2371",
		message:     "PART1",
	}
	deliverMsgPartCh <- deliverMsgPart{
		currentPart: 3,
		totalParts:  3,
		refNum:      109,
		sender:      "09191234567",
		receiver:    "2371",
		message:     "PART3",
	}
	deliverMsgPartCh <- deliverMsgPart{
		currentPart: 2,
		totalParts:  3,
		refNum:      109,
		sender:      "09191234567",
		receiver:    "2371",
		message:     "PART2",
	}
	actual := <-deliverMsgCompleteCh
	close(closeChan)
	wg.Wait()
	expectedMessage := "PART1PART2PART3"
	if actual.message != expectedMessage {
		t.Errorf("Expected %v, got %v\n", expectedMessage, actual.message)
	}
}

func TestReadCompleteDeliverMsg(t *testing.T) {
	wg := new(sync.WaitGroup)
	closeChan := make(chan struct{}, 1)
	deliverMsgCompleteCh := make(chan deliverMsgPart, 1)

	expectedSender := "09191234567"
	expectedReceiver := "2371"
	expectedMessage := "😃"

	f := func(sender, receiver, messageID, message, accessCode string) {
		if sender != expectedSender {
			t.Errorf("Expected %v, got %v\n", expectedSender, sender)
		}

		if receiver != expectedReceiver {
			t.Errorf("Expected %v, got %v\n", expectedReceiver, receiver)
		}

		if message != expectedMessage {
			t.Errorf("Expected %v, got %#v\n", expectedMessage, message)
		}
	}
	client := &Client{
		muconn: &sync.Mutex{},
		logger: log.New(os.Stdout, "debug ", 0),
	}
	readCompleteDeliveryMsg(wg, closeChan, deliverMsgCompleteCh, f, "", client)
	runtime.Gosched()
	deliverMsgCompleteCh <- deliverMsgPart{
		sender:   expectedSender,
		receiver: expectedReceiver,
		message:  "D83DDE03",
		dcs:      dcsXserUCS2,
	}
	close(closeChan)
	wg.Wait()
}

func TestEncodeDeliverMsg(t *testing.T) {

	testCases := []struct {
		expected string
		actual   string
	}{
		{
			expected: "Did you ever hear the tragedy of Darth Plagueis The Wise? I thought not. It's not a story the Jedi would tell you. It's a Sith legend. Darth Plagueis was",
			actual:   encodeDeliverMsg("44696420796F7520657665722068656172207468652074726167656479206F6620446172746820506C6167756569732054686520576973653F20492074686F75676874206E6F742E2049742773206E6F7420612073746F727920746865204A65646920776F756C642074656C6C20796F752E204974277320612053697468206C6567656E642E20446172746820506C61677565697320776173", dcsXserASCII),
		},
		{
			expected: "😃",
			actual:   encodeDeliverMsg("D83DDE03", dcsXserUCS2),
		},
	}
	for _, testCase := range testCases {

		if testCase.actual != testCase.expected {
			t.Errorf("Expected %s, got %s\n", testCase.expected, testCase.actual)
		}
	}
}
