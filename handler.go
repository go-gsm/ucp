package ucp

import "fmt"

type Handler func(sender, receiver, messageID, message, accessCode string)

// DefaultHandler is called if DeliveryHandler or ShortMessageHandler is not set.
// It just logs the parameters.
func DefaultHandler(sender, receiver, messageID, message, accessCode string) {
	fmt.Println("\nsender: ", sender, " receiver: ", receiver, " message_id: ", messageID, " message: ", message, " access_code: ", accessCode)
}
