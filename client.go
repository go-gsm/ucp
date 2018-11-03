package ucp

import (
	"bufio"
	"container/ring"
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Client represents a UCP client connection.
type Client struct {
	// addr represents the ip:port address of the SMSC.
	addr string
	// user represents the SMSC user.
	user string
	// password represents the SMSC password.
	password string
	// accessCode represents the SMSC access code.
	accessCode string
	// guards concurrent access to Client fields
	mu *sync.Mutex
	// billingID represents the billing identifier(similar to tariff class and service description in other protocols).
	billingID string
	// deliveryHandler is called whenever a delivery notification packet is received from the SMSC.
	deliveryHandler Handler
	// shortMessageHandler is called whenever a deliver short message packet is received from the SMSC.
	shortMessageHandler Handler
	// tps mobile-terminating transactions per second.
	tps int
	// muconn  guards concurrent access to net.Conn
	muconn *sync.Mutex
	// conn is the underlying network connection
	conn net.Conn
	// ringCounter is ringCounter buffer for the transaction reference numbers (00-99)
	ringCounter *ring.Ring
	// reader is a buffered reader used for reading incoming messages from the network.
	reader *bufio.Reader
	// writer is a buffered writer used for writing  packets to the network.
	writer *bufio.Writer
	// submitSmRespCh is a channel of submit sm response messages
	submitSmRespCh chan []string
	// deliverNotifCh is a channel of deliver notification messages
	deliverNotifCh chan []string
	// deliverMsgCh is a channel of deliver short messages (mobile-originating messages)
	deliverMsgCh chan []string
	// deliverMsgPartCh is a channel of incomplete mobile-originating multi-part messages
	deliverMsgPartCh chan deliverMsgPart
	// deliverMsgCompleteCh is a channel of completed mobile-originating multi-part messages
	deliverMsgCompleteCh chan deliverMsgPart
	// closeChan close channel
	closeChan chan struct{}
	// wg waitgroup for the goroutines
	wg *sync.WaitGroup
	// once is a sync.Once object to prevent closing the closeChan more than once
	once sync.Once
	// alertInterval is the interval for sending alert messages to the SMSC(ping messages)
	alertInterval time.Duration
	// rateLimiter is the rate limiter for sending mobile terminating messages
	rateLimiter *rate.Limiter
	// timeout network timeout for sending MTs, default is 5 seconds.
	timeout time.Duration
	// logger logs the debug messages
	logger Logger
}

// New returns a UCP client based on the given options.
func New(opt *Options) *Client {
	setDefaults(opt)
	return &Client{
		addr:                 opt.Addr,
		user:                 opt.User,
		password:             opt.Password,
		accessCode:           opt.AccessCode,
		tps:                  opt.Tps,
		submitSmRespCh:       make(chan []string, 1),
		deliverNotifCh:       make(chan []string, 1),
		deliverMsgCh:         make(chan []string, 1),
		deliverMsgPartCh:     make(chan deliverMsgPart, 1),
		deliverMsgCompleteCh: make(chan deliverMsgPart, 1),
		closeChan:            make(chan struct{}),
		alertInterval:        opt.KeepAlive,
		timeout:              opt.Timeout,
		deliveryHandler:      DefaultHandler,
		shortMessageHandler:  DefaultHandler,
		wg:                   new(sync.WaitGroup),
		logger:               opt.Logger,
		muconn:               new(sync.Mutex),
		mu:                   new(sync.Mutex),
	}
}

// Connect will attempt to establish a UCP connection to the SMSC.
func (c *Client) Connect() error {
	c.muconn.Lock()
	defer c.muconn.Unlock()
	c.initRefNum()
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)
	_, err = c.writer.Write(login(c.nextRefNum(), c.user, c.password))
	if err != nil {
		return err
	}
	err = c.writer.Flush()
	if err != nil {
		return err
	}
	resp, err := c.reader.ReadString(etx)
	if err != nil {
		return err
	}
	err = parseSessionResp(resp)
	if err != nil {
		return err
	}

	c.rateLimiter = rate.NewLimiter(rate.Limit(c.GetTps()), 1)
	sendAlert(c.nextRefNum(), c.user, c.writer, c.wg, c.closeChan, c.alertInterval, c.muconn, c)
	readLoop(c.reader, c.wg, c.closeChan, c.submitSmRespCh, c.deliverNotifCh, c.deliverMsgCh, c)
	readDeliveryNotif(c.writer, c.wg, c.closeChan, c.deliverNotifCh, c.deliveryHandler, c.accessCode, c.muconn, c)
	readDeliveryMsg(c.writer, c.wg, c.closeChan, c.deliverMsgCh, c.deliverMsgPartCh, c.deliverMsgCompleteCh, c.muconn, c)
	readPartialDeliveryMsg(c.wg, c.closeChan, c.deliverMsgPartCh, c.deliverMsgCompleteCh, c)
	readCompleteDeliveryMsg(c.wg, c.closeChan, c.deliverMsgCompleteCh, c.shortMessageHandler, c.accessCode, c)
	return err
}

// initRefNum initializes the ringCounter counter from 00 to 99
func (c *Client) initRefNum() {
	ringCounter := ring.New(maxRefNum)
	for i := 0; i < maxRefNum; i++ {
		ringCounter.Value = []byte(fmt.Sprintf("%02d", i))
		ringCounter = ringCounter.Next()
	}
	c.ringCounter = ringCounter
}

// nextRefNum returns the next transaction reference number
func (c *Client) nextRefNum() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	refNum := (c.ringCounter.Value).([]byte)
	c.ringCounter = c.ringCounter.Next()
	c.Printf("transaction reference number: %s\n", refNum)
	return refNum
}

// GetTps returns the mobile-terminating transactions per second.
func (c *Client) GetTps() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tps
}

// SetTps sets the mobile-terminating transactions per second.
func (c *Client) SetTps(tps int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tps = tps
}

// SetBillingID sets the billing identifier to be used by the client.
func (c *Client) SetBillingID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.billingID = id
}

// GetBillingID returns the current billing identifier.
func (c *Client) GetBillingID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.billingID
}

// DeliveryHandler sets the delivery notification handler.
func (c *Client) DeliveryHandler(handler Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deliveryHandler = handler
}

// ShortMessageHandler sets the delivery short message handler.
func (c *Client) ShortMessageHandler(handler Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shortMessageHandler = handler
}

// Send will send the message to the receiver with a sender mask.
// It returns a list of message IDs from the SMSC.
func (c *Client) Send(sender, receiver, message string) ([]string, error) {
	c.muconn.Lock()
	defer c.muconn.Unlock()

	msgType := getMessageType(message)
	msgParts := getMessageParts(message)
	refNum := rand.Intn(maxRefNum)
	ids := make([]string, len(msgParts))
	c.rateLimiter.SetLimit(rate.Limit(c.GetTps()))
	for i := 0; i < len(msgParts); i++ {
		sendPacket := encodeMessage(c.nextRefNum(), sender, receiver, msgParts[i], msgType,
			c.GetBillingID(), refNum, i+1, len(msgParts))
		c.rateLimiter.Wait(context.Background())
		c.Printf("sendPacket: %q\n", sendPacket)
		if _, err := c.writer.Write(sendPacket); err != nil {
			c.Printf("error writing sendPacket: %v\n", err)
			return ids, err
		}
		if err := c.writer.Flush(); err != nil {
			c.Printf("error flushing sendPacket: %v\n", err)
			return ids, err
		}
		select {
		case fields := <-c.submitSmRespCh:
			ack := fields[ackIndex]
			if ack == negativeAck {
				errMsg := fields[len(fields)-errMsgOffset]
				errCode := fields[len(fields)-errCodeOffset]
				c.Printf("negative ack, errMsg: %v errCode: %v\n", errMsg, errCode)
				return ids, &UcpError{errCode, errMsg}
			}
			id := fields[submitSmIdIndex]
			ids[i] = id
		case <-time.After(c.timeout):
			c.Printf("send timeout\n")
			return ids, &UcpError{errCodeTimeout, "Network time-out"}
		}
	}
	return ids, nil
}

// Close will close the UCP connection
func (c *Client) Close() {
	c.Printf("closing client\n")

	// signal all the goroutines to exit
	// guarantee that the close channel will only be closed once
	c.once.Do(func() {
		c.Printf("closing closeChan\n")
		close(c.closeChan)
	})

	// close the TCP connection
	if c.conn != nil {
		c.Printf("closing tcp connection\n")
		c.conn.Close()
	}

	// wait for all the pending goroutines to exit gracefully
	c.wg.Wait()
	c.Printf("closed client successfully\n")
}
