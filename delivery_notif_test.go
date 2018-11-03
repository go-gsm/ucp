package ucp

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"testing"
)

func TestDeliveryNotifAckPacket(t *testing.T) {
	actual := deliveryNotifAckPacket([]byte("03"), "")
	expected := []byte("\x0203/00020/R/53/A///99\x03")
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected %s, got %s\n", expected, actual)
	}
}

func TestReadDeliveryNotif(t *testing.T) {
	buf := new(bytes.Buffer)
	writer := bufio.NewWriter(buf)
	wg := new(sync.WaitGroup)
	closeChan := make(chan struct{}, 1)
	deliverNotifCh := make(chan []string, 1)

	expectedSender := "2371"
	expectedReceiver := "09191234567"
	expectedMessage := "Message for +639191234567, with identification 170911160250 has been delivered on 2017-09-11 at 16:02:52."
	f := func(sender, receiver, messageID, message, accessCode string) {
		if sender != expectedSender {
			t.Errorf("Expected %v, got %v\n", expectedSender, sender)
		}

		if receiver != expectedReceiver {
			t.Errorf("Expected %v, got %v\n", expectedReceiver, receiver)
		}

		if message != expectedMessage {
			fmt.Printf("Expected %v, got %#v\n", expectedMessage, message)
		}
	}
	client := &Client{
		muconn: &sync.Mutex{},
		logger: log.New(os.Stdout, "debug ", 0),
	}
	readDeliveryNotif(writer, wg, closeChan, deliverNotifCh, f, "",
		client.muconn, client)
	runtime.Gosched()
	deliverNotifCh <- []string{"00", "00304", "O", "53", "2371", "09191234567", "", "", "", "", "", "", "", "", "", "", "", "", "110917160250", "0", "000", "110917160252", "3", "", "4D65737361676520666F72202B3633393139313233343536372C2077697468206964656E74696669636174696F6E2031373039313131363032353020686173206265656E2064656C697665726564206F6E20323031372D30392D31312061742031363A30323A35322E", "1", "", "", "", "", "", "", "", "", "", "", "", "91"}

	close(closeChan)
	wg.Wait()
	actual := buf.Bytes()
	expected := []byte("\x0200/00044/R/53/A//09191234567:110917160250/76\x03")

	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected %s, got %s\n", expected, actual)
	}

}
