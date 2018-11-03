package ucp

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"unicode/utf8"

	"github.com/go-gsm/charset"
)

// encode returns [][]byte that represents
// the list of struct fields(as bytes) of a packet
func encode(i operation) [][]byte {
	fields := reflect.TypeOf(i)
	values := reflect.ValueOf(i)
	num := fields.NumField()
	data := make([][]byte, num)
	for j := 0; j < num; j++ {
		value := values.Field(j)
		data[j] = value.Bytes()
	}
	return data
}

// getMessageType returns the message type.
// alphaNumericMessage for ASCII messages,
// transparentData for Unicode messages.
func getMessageType(message string) string {
	if charset.IsGsmAlpha(message) {
		return alphaNumericMessage
	}
	return transparentData
}

// getMessageParts splits the message into a list
// such that the length of each element of the string slice
// is less than the allowed maximum length
func getMessageParts(message string) []string {
	if charset.IsGsmAlpha(message) {
		return asciiParts(message)
	}
	return ucs2Parts(message)
}

// asciiParts splits ASCII messages
func asciiParts(message string) []string {
	// less than max, no need to split
	if len(charset.Encode7Bit(message)) <= gsmMaxSinglePart {
		return []string{message}
	}
	// greater than max, split it.
	// gsmMaxMultiPart is 153 bcoz we need to use the length of the UDH field (in octets),
	// multiplied by 8/7, rounded up to the nearest integer value.
	//
	// length of the UDH field (in octets) => 6
	// 6 * (8/7) => 7 (rounded up)
	// 160 - 7 = 153
	parts := make([]string, 0)
	for start, end := 0, gsmMaxMultiPart; len(message) > start; start, end = start+gsmMaxMultiPart, end+gsmMaxMultiPart {
		if len(message) < end {
			end = len(message)
		}
		part := message[start:end]
		parts = append(parts, part)

	}
	return parts
}

// ucs2Parts splits Unicode messages
func ucs2Parts(message string) []string {
	// less than max, no need to split
	if len(message) <= ucs2MaxSinglePart {
		return []string{message}
	}
	// greater than max, split it.
	// ucs2MaxMultiPart is 64 bcoz we need to use the length of the UDH field (in octets).
	//  The length of the UDH field (in octets) plus the length of the TMsg field (in octets) must not exceed 140.
	// 140 octets is equal to 70 hextets(UCS-2)
	// 70 - 6 = 64
	parts := make([]string, 0)
	for start, end := 0, ucs2MaxMultiPart; len(message) > start; start, end = end, end+ucs2MaxMultiPart {
		if len(message) < end {
			end = len(message)
		}
		part := message[start:end]
		for !utf8.ValidString(part) {
			// part is not valid, backoff by one
			end -= 1
			// re-slice the message
			part = message[start:end]
		}
		parts = append(parts, part)
	}
	return parts
}

func getDataCodingScheme(msgType string) string {
	if msgType == alphaNumericMessage {
		return dataCodingSchemeASCII
	}
	return dataCodingSchemeUCS2
}

// buildHexMsg formats the message to a hex string
func buildHexMsg(messageType string, message string) string {
	if messageType == alphaNumericMessage {
		return fmt.Sprintf("%02X", string(charset.Encode7Bit(message)))
	}
	return fmt.Sprintf("%04X", string(charset.EncodeUcs2(message)))
}

// maskSender formats the sender mask to a hex string
func maskSender(sender string) string {
	encodedSender := charset.Pack7Bit(charset.Encode7Bit(sender))
	encodedSenderSize := len(encodedSender) * 2
	senderByteSlice := make([]byte, 0)
	senderByteSlice = append(senderByteSlice, []byte{byte(encodedSenderSize)}...)
	senderByteSlice = append(senderByteSlice, encodedSender...)
	encodedHexSender := fmt.Sprintf("%02X", string(senderByteSlice))
	return encodedHexSender
}

// encodeMessage builds a submit sm packet
func encodeMessage(transRefNum []byte, sender, receiver, message, messageType, billingID string,
	referenceNum, msgPartNum, totalMsgParts int) []byte {

	encodedHexSender := maskSender(sender)
	encodedHexMessage := buildHexMsg(messageType, message)
	numBits := strconv.Itoa(len(encodedHexMessage) * 4)
	xserData := buildXser(billingID, messageType, referenceNum, totalMsgParts, msgPartNum)

	s := submit{
		AdC:  []byte(receiver),
		OAdC: []byte(encodedHexSender),
		NRq:  []byte(nAdCUsed),
		NT:   []byte(notificationTypeDN),
		MT:   []byte(messageType),
		NB:   []byte(numBits),
		Msg:  []byte(encodedHexMessage),
		MCLs: []byte(messageClass),
		OTOA: []byte(oAdCAlphaNum),
		Xser: []byte(xserData),
	}

	buf := preparePacket(transRefNum, s)
	return buf
}

// preparePacket builds a packet to be written, complete with its length and checksum
func preparePacket(transRefNum []byte, p operation) []byte {
	buf := make([]byte, 0)
	buf = append(buf, stx)
	fields := encode(p)
	joinedFields := bytes.Join(fields, []byte(delimiter))
	Len := pduLenMinusData + len(joinedFields)
	partial := [][]byte{
		transRefNum,
		[]byte(fmt.Sprintf("%05d", Len)),
		p.Type(),
		p.Code(),
		joinedFields,
	}
	pduWithoutChecksum := append(bytes.Join(partial, []byte(delimiter)), []byte(delimiter)...)
	pduWithChecksum := append(pduWithoutChecksum, checksum(pduWithoutChecksum)...)
	buf = append(buf, pduWithChecksum...)
	buf = append(buf, etx)
	return buf
}
