package ucp

import "fmt"

// UCP protocol error
type UcpError struct {
	Code string
	Msg  string
}

func (e *UcpError) Error() string {
	return fmt.Sprintf("[ucp error] code: %s message: %s", e.Code, e.Msg)
}
