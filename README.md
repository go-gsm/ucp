[![Build Status](https://travis-ci.org/go-gsm/ucp.svg?branch=master)](https://travis-ci.org/go-gsm/ucp) [![GoDoc](https://godoc.org/github.com/go-gsm/ucp?status.svg)](https://godoc.org/github.com/go-gsm/ucp) [![Coverage Status](https://coveralls.io/repos/github/go-gsm/ucp/badge.svg?branch=master)](https://coveralls.io/github/go-gsm/ucp?branch=master)[![Go Report Card](https://goreportcard.com/badge/github.com/go-gsm/ucp)](https://goreportcard.com/report/github.com/go-gsm/ucp)

## ucp

`ucp` is a pure [Go](https://golang.org) implementation of the [UCP](https://wiki.wireshark.org/UCP) protocol primarily used to connect to short message service centres (SMSCs),  in order to send and receive short messages (SMS).

#### setup
- go 1.11
- git

#### installation
```
go get github.com/go-gsm/ucp
```

#### usage
```
opt := &ucp.Options{
  Addr:       SMSC_ADDR,
  User:       SMSC_USER,
  Password:   SMSC_PASSWORD,
  AccessCode: SMSC_ACCESSCODE,
}
client := ucp.New(opt)
client.Connect()
defer client.Close()
ids, err := client.Send(sender, receiver, message)
```

#### demo

[ucp-cli](https://github.com/go-gsm/ucp-cli)

![demo](
https://thumbs.gfycat.com/HorribleWelcomeAcouchi-size_restricted.gif)


