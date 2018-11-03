package ucp

import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/time/rate"
	"log"
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestInitRefNum(t *testing.T) {
	client := &Client{
		mu:     &sync.Mutex{},
		logger: log.New(os.Stdout, "debug ", 0),
	}
	client.initRefNum()
	firstRefNum := client.nextRefNum()
	expectedFirstRefNum := []byte("00")
	if !bytes.Equal(firstRefNum, expectedFirstRefNum) {
		t.Errorf("Expected %v got %v\n", expectedFirstRefNum, firstRefNum)
	}
	// advance to "99"
	for i := 1; i <= 99; i++ {
		client.nextRefNum()
	}
	// should reset to "00"
	firstRefNum = client.nextRefNum()
	if !bytes.Equal(firstRefNum, expectedFirstRefNum) {
		t.Errorf("Expected %v got %v\n", expectedFirstRefNum, firstRefNum)
	}
}

func TestSend(t *testing.T) {
	buf := new(bytes.Buffer)
	submitSmRespCh := make(chan []string, 1)
	client := &Client{
		mu:             &sync.Mutex{},
		muconn:         &sync.Mutex{},
		logger:         log.New(os.Stdout, "debug ", 0),
		rateLimiter:    rate.NewLimiter(rate.Limit(1), 1),
		writer:         bufio.NewWriter(buf),
		submitSmRespCh: submitSmRespCh,
	}
	ack := []string{"01", "00044", "R", "51", "A", "", "09191234567:110917173639", "95"}
	client.initRefNum()

	select {
	case submitSmRespCh <- ack:
		fmt.Println("sent ack")
	default:
		fmt.Println("cant send ack")
	}

	ids, err := client.Send("test", "09191234567", "hello world")
	expectedIds := []string{"09191234567:110917173639"}
	expectedBytesWritten := []byte("\x0200/00120/O/51/09191234567/08F4F29C0E//1//1/////////////3/88/68656C6C6F20776F726C64////1////5039//020100060101070101///B7\x03")
	actualBytesWritten := buf.Bytes()

	if !bytes.Equal(expectedBytesWritten, actualBytesWritten) {
		t.Errorf("Expected %v got %v\n", expectedBytesWritten, actualBytesWritten)
	}

	if err != nil {
		t.Error("Expected nil error\n")
	}

	if !reflect.DeepEqual(expectedIds, ids) {
		t.Errorf("Expected %v got %v\n", expectedIds, ids)
	}

}
