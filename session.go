package ucp

import "fmt"

type session struct {
	OAdC []byte
	OTON []byte
	ONPI []byte
	STYP []byte
	PWD  []byte
	NPWD []byte
	VERS []byte
	LAdC []byte
	LTON []byte
	LNPI []byte
	OPID []byte
	RES1 []byte
}

func (s session) Code() []byte {
	return []byte(opSessionManagement)
}

func (s session) Type() []byte {
	return []byte(operationType)
}

// login creates a packet to be used for session management
func login(transRefNum []byte, user string, password string) []byte {
	encodedPassword := fmt.Sprintf("%02X", password)
	s := session{
		OAdC: []byte(user),
		OTON: []byte(abbreviatedNumber),
		ONPI: []byte(smscSpecific),
		STYP: []byte(openSession),
		PWD:  []byte(encodedPassword),
		VERS: []byte(vers),
	}
	buf := preparePacket(transRefNum, s)
	return buf
}
