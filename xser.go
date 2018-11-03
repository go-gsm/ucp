package ucp

import (
	"bytes"
	"fmt"
	"strconv"
)

// buildXSerUDH builds a user data header extra service packet.
func buildXSerUDH(referenceNum, totalMsgParts, msgPartNum int) string {
	// only one message part, no need for UDH
	if totalMsgParts == 1 {
		return ""
	}
	return fmt.Sprintf("%s%02X%02X%02X", concatMsgTLDD, referenceNum, totalMsgParts, msgPartNum)
}

// buildXSERBillingID builds a billing identifier extra service packet
func buildXSERBillingID(billingID string) string {
	billingIDPacketLen := len(billingID)
	if billingIDPacketLen == 0 {
		return ""
	}
	return fmt.Sprintf("%s%02X%02X", billingIDXserKey, billingIDPacketLen, billingID)
}

// parseXser returns a map of xser identifier to xser data from a given xser packet.
func parseXser(xser string) map[string]string {
	m := make(map[string]string)
	if xser == "" {
		return m
	}
	buf := bytes.NewBufferString(xser)
	for buf.Len() > 0 {
		xserType := buf.Next(2)
		xserLen := buf.Next(2)
		convXserlen, err := strconv.ParseInt(string(xserLen), 16, 0)
		if err != nil {
			// return an empty map indicating that an xser key-value pair cant be extracted from the string
			return m
		}
		xserData := buf.Next(int(convXserlen) * 2)
		m[string(xserType)] = string(xserData)
	}
	return m
}

// buildXser builds all the required extra services together.
func buildXser(billingID, messageType string, referenceNum, totalMsgParts, msgPartNum int) string {
	xserData := getDataCodingScheme(messageType) +
		buildXSerUDH(referenceNum, totalMsgParts, msgPartNum) +
		buildXSERBillingID(billingID) +
		urgencyIndicatorNormal +
		ackReqDeliveryAck
	return xserData
}
