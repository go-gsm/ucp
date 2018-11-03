package ucp

import (
	"errors"
	"strings"
)

var errInvalidPacket = errors.New("invalid packet")

// parseSessionResp parse the initial login
func parseSessionResp(resp string) error {
	removedStxEtx := strings.TrimFunc(resp, func(c rune) bool {
		return c == stx || c == etx
	})
	splitFields := strings.Split(removedStxEtx, delimiter)
	if len(splitFields) < openSesRespMinLen {
		return errInvalidPacket
	}
	opType := splitFields[optypeIndex]
	ack := splitFields[ackIndex]
	if opType != opSessionManagement {
		return errInvalidPacket
	}
	if ack != positiveAck {
		errMsg := splitFields[len(splitFields)-errMsgOffset]
		errCode := splitFields[len(splitFields)-errCodeOffset]
		return &UcpError{errCode, errMsg}
	}
	return nil
}

func parseResp(resp string) (string, []string, error) {
	removedStxEtx := strings.TrimFunc(resp, func(c rune) bool {
		return c == stx || c == etx
	})
	splitFields := strings.Split(removedStxEtx, delimiter)
	if len(splitFields) < respMinLen {
		return "", []string{}, errInvalidPacket
	}
	opType := splitFields[optypeIndex]

	return opType, splitFields, nil
}
