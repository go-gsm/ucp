package ucp

import (
	"bufio"
	"sync"
	"time"
)

type alert struct {
	AdC []byte
	PID []byte
}

func (a alert) Code() []byte {
	return []byte(opAlert)
}

func (a alert) Type() []byte {
	return []byte(operationType)
}

// ping returns a []byte to send as a keep-alive
func ping(transRefNum []byte, user string) []byte {
	a := alert{
		AdC: []byte(user),
		PID: []byte(pcAppOverTcpIp),
	}
	buf := preparePacket(transRefNum, a)
	return buf
}

// sendAlert writes an alert packet to the socket every alertInterval.
func sendAlert(transRefNum []byte, user string, writer *bufio.Writer, wg *sync.WaitGroup,
	closeChan chan struct{}, alertInterval time.Duration, mu *sync.Mutex, logger Logger) {
	wg.Add(1)
	ticker := time.NewTicker(alertInterval)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-closeChan:
				ticker.Stop()
				logger.Printf("stopped ticker\n")
				logger.Printf("sendAlert terminated\n")
				return
			case <-ticker.C:
				mu.Lock()
				if _, err := writer.Write(ping(transRefNum, user)); err != nil {
					logger.Printf("error writing ping: %v\n", err)
				}
				if err := writer.Flush(); err != nil {
					logger.Printf("error flushing ping: %v\n", err)
				}
				mu.Unlock()
			}
		}
	}()

}
