package ucp

import (
	"bufio"
	"io"
	"sync"
)

// readLoop reads incoming messages from the SMSC using the underlying bufio.Reader
func readLoop(reader *bufio.Reader, wg *sync.WaitGroup, closeChan chan struct{},
	submitSmRespCh, deliverNotifCh, deliverMsgCh chan []string, logger Logger) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-closeChan:
				logger.Printf("readLoop terminated\n")
				return
			default:
				readData, err := reader.ReadString(etx)
				if err != nil {
					if err == io.EOF {
						logger.Printf("read EOF\n")
						return
					}
					continue
				}
				opType, fields, err := parseResp(readData)
				if err != nil {
					continue
				}
				switch opType {
				case opSubmitShortMessage:
					logger.Printf("opSubmitShortMessage: %q\n", fields)
					submitSmRespCh <- fields
				case opDeliveryNotification:
					logger.Printf("opDeliveryNotification: %q\n", fields)
					deliverNotifCh <- fields
				case opDeliveryShortMessage:
					logger.Printf("opDeliveryShortMessage: %q\n", fields)
					deliverMsgCh <- fields
				case opAlert:
					logger.Printf("opAlert: %q\n", fields)
				default:
					logger.Printf("unknown operationType: %q\n", fields)
				}
			}
		}
	}()
}
