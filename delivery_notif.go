package ucp

import (
	"bufio"
	"encoding/hex"
	"sync"
)

type deliveryNotification struct {
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

// deliveryNotifAck represents a data structure of an acknowledgment packet for deliver sm and deliver notif.
type deliveryNotifAck struct {
	Ack                    []byte
	ModifiedValidityPeriod []byte
	SystemMessage          []byte
}

func (s deliveryNotifAck) Code() []byte {
	return []byte(opDeliveryNotification)
}

func (s deliveryNotifAck) Type() []byte {
	return []byte(resultType)
}

// readDeliveryNotif reads all deliver notifications from deliverNotifCh channel.
// Once a deliver notification message is read, it sends an ack to the SMSC and
// calls deliveryHandler.
func readDeliveryNotif(writer *bufio.Writer, wg *sync.WaitGroup, closeChan chan struct{},
	deliverNotifCh chan []string, deliveryHandler Handler, accessCode string, mu *sync.Mutex, logger Logger) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-closeChan:
				logger.Printf("readDeliveryNotif terminated\n")
				return
			case dr := <-deliverNotifCh:
				refNum := dr[refNumIndex]
				msg, _ := hex.DecodeString(dr[drMsgIndex])
				sender := dr[drSenderIndex]
				recvr := dr[drRecvrIndex]
				scts := dr[drSctsIndex]
				msgID := recvr + ":" + scts
				mu.Lock()
				if _, err := writer.Write(deliveryNotifAckPacket([]byte(refNum), msgID)); err != nil {
					logger.Printf("error writing delivery notification ack packet: %v\n", err)
				}
				if err := writer.Flush(); err != nil {
					logger.Printf("error flushing delivery notification ack packet: %v\n", err)
				}
				mu.Unlock()
				deliveryHandler(sender, recvr, msgID, string(msg), accessCode)
			}
		}
	}()
}

// deliveryAckPDU builds a deliveryNotifAck packet
func deliveryNotifAckPacket(refNum []byte, systemMessage string) []byte {
	ack := deliveryNotifAck{
		Ack:           []byte(positiveAck),
		SystemMessage: []byte(systemMessage),
	}
	buf := preparePacket(refNum, ack)
	return buf
}
