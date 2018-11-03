package ucp

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/go-gsm/charset"
)

type deliverShortMessage struct {
	AdC   []byte
	OAdC  []byte
	AC    []byte
	NRq   []byte
	NAdC  []byte
	NT    []byte
	NPID  []byte
	LRq   []byte
	LRAd  []byte
	LPID  []byte
	DD    []byte
	DDT   []byte
	VP    []byte
	RPID  []byte
	SCTS  []byte
	Dst   []byte
	Rsn   []byte
	DSCTS []byte
	MT    []byte
	NB    []byte
	Msg   []byte
	MMS   []byte
	PR    []byte
	DCs   []byte
	MCLs  []byte
	RPI   []byte
	CPg   []byte
	RPLy  []byte
	OTOA  []byte
	HPLMN []byte
	Xser  []byte
	RES4  []byte
	RES5  []byte
}

// deliverySmAck represents a data structure of an acknowledgment packet for deliver sm.
type deliverySmAck struct {
	Ack                    []byte
	ModifiedValidityPeriod []byte
	SystemMessage          []byte
}

func (s deliverySmAck) Code() []byte {
	return []byte(opDeliveryShortMessage)
}

func (s deliverySmAck) Type() []byte {
	return []byte(resultType)
}

// deliveryAckPDU builds a deliveryNotifAck packet
func deliverySmAckPacket(refNum []byte, systemMessage string) []byte {
	ack := deliverySmAck{
		Ack:           []byte(positiveAck),
		SystemMessage: []byte(systemMessage),
	}
	buf := preparePacket(refNum, ack)
	return buf
}

// readDeliveryMsg reads all deliver sm messages(mobile-originating messages) from the deliverMsgCh channel.
func readDeliveryMsg(writer *bufio.Writer, wg *sync.WaitGroup, closeChan chan struct{},
	deliverMsgCh chan []string, deliverMsgPartCh, deliverMsgCompleteCh chan deliverMsgPart, mu *sync.Mutex, logger Logger) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-closeChan:
				logger.Printf("readDeliveryMsg terminated\n")
				return
			case mo := <-deliverMsgCh:
				xser := mo[xserIndex]
				xserData := parseXser(xser)
				msg := mo[moMsgIndex]
				refNum := mo[refNumIndex]
				sender := mo[moSenderIndex]
				recvr := mo[moRecvrIndex]
				scts := mo[moSctsIndex]
				sysmsg := recvr + ":" + scts
				msgID := sender + ":" + scts

				mu.Lock()
				// send ack to SMSC with the same reference number
				if _, err := writer.Write(deliverySmAckPacket([]byte(refNum), sysmsg)); err != nil {
					logger.Printf("error writing delivery sm ack packet: %v\n", err)
				}
				if err := writer.Flush(); err != nil {
					logger.Printf("error flushing delivery sm ack packet: %v\n", err)
				}
				mu.Unlock()

				var incomingMsg deliverMsgPart
				incomingMsg.sender = sender
				incomingMsg.receiver = recvr
				incomingMsg.message = msg
				incomingMsg.msgID = msgID

				if xserDCS, ok := xserData[dcsXserKey]; ok {
					incomingMsg.dcs = xserDCS
				}

				// check the user data header extra service field
				// if it exists, the incoming message has multiple parts
				if xserUdh, ok := xserData[udhXserKey]; ok {
					// handle multipart mobile originating message i.e. len(message) > 140 bytes
					// get the total message parts in the xser data
					msgPartsLen := xserUdh[len(xserUdh)-4 : len(xserUdh)-2]
					// get the current message part in the xser data
					msgPart := xserUdh[len(xserUdh)-2:]
					// get UDH reference number
					msgRefNum := xserUdh[len(xserUdh)-6 : len(xserUdh)-4]
					// convert hexstring to integer
					msgRefNumInt, _ := strconv.ParseInt(msgRefNum, 16, 0)
					msgPartsLenInt, _ := strconv.ParseInt(msgPartsLen, 16, 64)
					msgPartInt, _ := strconv.ParseInt(msgPart, 16, 64)
					// convert int64 to int
					incomingMsg.currentPart = int(msgPartInt)
					incomingMsg.totalParts = int(msgPartsLenInt)
					incomingMsg.refNum = int(msgRefNumInt)
					// send to partial channel

					deliverMsgPartCh <- incomingMsg

				} else {
					// handle mobile originating message with only 1 part i.e. len(message) <= 140 bytes
					// send the incoming message to the complete channel
					deliverMsgCompleteCh <- incomingMsg

				}
			}

		}
	}()
}

// readPartialDeliveryMsg concatenates partial incoming mobile-originating messages
func readPartialDeliveryMsg(wg *sync.WaitGroup, closeChan chan struct{},
	deliverMsgPartCh, deliverMsgCompleteCh chan deliverMsgPart, logger Logger) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		concatMap := make(map[string][]deliverMsgPart)
		for {
			select {
			case <-closeChan:
				logger.Printf("readPartialDeliveryMsg terminated\n")
				return
			case partial := <-deliverMsgPartCh:
				mapKey := fmt.Sprintf("%s:%s:%d", partial.sender, partial.receiver, partial.refNum)
				partMsgList, ok := concatMap[mapKey]
				if ok {
					partMsgList = append(partMsgList, partial)
					concatMap[mapKey] = partMsgList
					if len(partMsgList) == partial.totalParts {
						sort.Slice(partMsgList, func(i, j int) bool {
							return partMsgList[i].currentPart < partMsgList[j].currentPart
						})
						var fullMsg string
						for _, partMsg := range partMsgList {
							fullMsg += partMsg.message
						}
						partial.message = fullMsg
						deliverMsgCompleteCh <- partial
						delete(concatMap, mapKey)
					}
				} else {
					concatMap[mapKey] = []deliverMsgPart{partial}
				}

			}
		}
	}()
}

// readPartialDeliveryMsg processes complete incoming mobile-originating messages
func readCompleteDeliveryMsg(wg *sync.WaitGroup, closeChan chan struct{},
	deliverMsgCompleteCh chan deliverMsgPart, shortMessageHandler Handler, accessCode string, logger Logger) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-closeChan:
				logger.Printf("readCompleteDeliveryMsg terminated\n")
				return
			case complete := <-deliverMsgCompleteCh:
				shortMessageHandler(
					complete.sender,
					complete.receiver,
					complete.msgID,
					encodeDeliverMsg(complete.message, complete.dcs),
					accessCode,
				)
			}
		}
	}()
}

// encodeDeliverMsg encodes a mobile-originating message
func encodeDeliverMsg(mo string, dcs string) string {
	var msg string
	if dcs == dcsXserASCII {
		msgByte, _ := charset.ParseOddHexStr(mo)
		msg, _ = charset.Decode7Bit(msgByte)
	}
	if dcs == dcsXserUCS2 {
		decoded, _ := hex.DecodeString(mo)
		msg, _ = charset.DecodeUcs2(decoded)
	}
	return msg
}

// deliverMsgPart represents a deliver sm message part
type deliverMsgPart struct {
	currentPart int
	totalParts  int
	refNum      int
	sender      string
	receiver    string
	message     string
	msgID       string
	dcs         string
}
