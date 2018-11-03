package ucp

import (
	"fmt"
)

// checksum computes the checksum of a UCP packet.
func checksum(b []byte) []byte {
	var sum byte
	for _, i := range b {
		sum += i
	}
	mask := sum & 0xFF
	chksum := fmt.Sprintf("%02X", mask)
	return []byte(chksum)
}
