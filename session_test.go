package ucp

import (
	"bytes"
	"testing"
)

func TestSession(t *testing.T) {
	actual := login([]byte("00"), "emi_client", "password")
	data := struct {
		actual   []byte
		expected []byte
	}{
		actual,
		[]byte("\x0200/00061/O/60/emi_client/6/5/1/70617373776F7264//0100//////D1\x03"),
	}

	if !bytes.Equal(data.expected, data.actual) {
		t.Errorf("Expected %v, got %v\n", string(data.expected[:]), string(data.actual[:]))
	}
}
