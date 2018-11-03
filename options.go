package ucp

import "time"

// Options is used to configure the instantiation of a Client.
type Options struct {
	// Addr represents the ip:port address of the SMSC.
	Addr string
	// User represents the SMSC user.
	User string
	// Password represents the SMSC password.
	Password string
	// AccessCode represents the SMSC access code.
	AccessCode string
	// Tps mobile-terminating transactions per second.
	Tps int
	// Logger implements the Logger interface
	Logger Logger
	// Timeout is the specified network timeout for waiting submit short message responses
	Timeout time.Duration
	// KeepAlive is the ping interval for sending keep-alive packets to the SMSC
	KeepAlive time.Duration
	// DeliveryHandler sets the delivery notification handler(delivery receipts).
	DeliveryHandler Handler
	// ShortMessageHandler sets the delivery short message handler(mobile originating messages).
	ShortMessageHandler Handler
}

func setDefaults(opt *Options) *Options {
	if opt.Tps == 0 {
		opt.Tps = 10
	}
	if opt.KeepAlive == 0 {
		opt.KeepAlive = 30 * time.Second
	}
	if opt.Timeout == 0 {
		opt.Timeout = 5 * time.Second
	}
	if opt.DeliveryHandler == nil {
		opt.DeliveryHandler = DefaultHandler
	}
	if opt.ShortMessageHandler == nil {
		opt.ShortMessageHandler = DefaultHandler
	}
	return opt
}
