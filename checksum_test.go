package ucp

import (
	"bytes"
	"testing"
)

func TestChecksum(t *testing.T) {
	data := struct {
		exp    []byte
		packet []byte
	}{
		[]byte("61"),
		[]byte("02/00059/O/60/07656765/2/1/1/50617373776F7264//0100//////"),
	}
	out := checksum(data.packet)
	if !bytes.Equal(data.exp, out) {
		t.Errorf("Expected %v, got %v\n", data.exp, out)
	}
}
