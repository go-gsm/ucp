package ucp

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestAlert(t *testing.T) {

	actual := ping([]byte("01"), "emi_client")
	data := struct {
		actual   []byte
		expected []byte
	}{
		actual,
		[]byte("\x0201/00032/O/31/emi_client/0539/0D\x03"),
	}

	if !bytes.Equal(data.expected, data.actual) {
		t.Errorf("Expected %v, got %v\n", string(data.expected[:]), string(data.actual[:]))
	}
}

func TestSendAlert(t *testing.T) {
	buf := new(bytes.Buffer)
	writer := bufio.NewWriter(buf)
	wg := new(sync.WaitGroup)
	closeChan := make(chan struct{}, 1)
	mu := new(sync.Mutex)
	sendAlert([]byte("01"), "emi_client", writer, wg, closeChan, 500*time.Millisecond, mu,
		&Client{logger: log.New(os.Stdout, "debug ", 0)})
	runtime.Gosched()
	time.Sleep(700 * time.Millisecond)
	close(closeChan)
	wg.Wait()
	expected := []byte("\x0201/00032/O/31/emi_client/0539/0D\x03")
	actual := buf.Bytes()
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected %v, got %v\n", string(expected[:]), string(actual[:]))
	}
}
