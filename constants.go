package ucp

const (

	// pduLenMinusData the length of the PDU minus the encoded data
	// +--------------------------------------------------------------------------+
	// | X | X | / | X | X | X | X | X | / | X | / | X | X | / | DATA | / | X | X |
	// | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10| 11| 12| 13| 14|      | 15| 16| 17|
	// +--------------------------------------------------------------------------+
	pduLenMinusData                = 17
	stx                            = 2
	etx                            = 3
	delimiter                      = "/"
	opAlert                        = "31"
	opSubmitShortMessage           = "51"
	opDeliveryShortMessage         = "52"
	opDeliveryNotification         = "53"
	opSessionManagement            = "60"
	positiveAck                    = "A"
	negativeAck                    = "N"
	vers                           = "0100"
	abbreviatedNumber              = "6"
	smscSpecific                   = "5"
	openSession                    = "1"
	pcAppOverTcpIp                 = "0539"
	oAdCAlphaNum                   = "5039"
	nAdCUsed                       = "1"
	notificationTypeDN             = "1"
	messageClass                   = "1"
	operationType                  = "O"
	resultType                     = "R"
	numericMessage                 = "2"
	alphaNumericMessage            = "3"
	transparentData                = "4"
	concatMsgTLDD                  = "0106050003"
	urgencyIndicatorBulk           = "060100"
	urgencyIndicatorNormal         = "060101"
	urgencyIndicatorUrgent         = "060102"
	dataCodingSchemeASCII          = "020100"
	dataCodingSchemeUCS2           = "020108"
	dcsXserASCII                   = "00"
	dcsXserUCS2                    = "08"
	ackReqNoAck                    = "070100"
	ackReqDeliveryAck              = "070101"
	ackRequestManualAck            = "070102"
	ackRequestDeliveryAndManualAck = "070103"
	udhXserKey                     = "01"
	billingIDXserKey               = "0C"
	dcsXserKey                     = "02"
	optypeIndex                    = 3
	openSesRespMinLen              = 6
	respMinLen                     = 4
	maxRefNum                      = 100
	submitSmIdIndex                = 6
	gsmMaxSinglePart               = 160
	gsmMaxMultiPart                = 153
	ucs2MaxSinglePart              = 70
	ucs2MaxMultiPart               = 64
	refNumIndex                    = 0
	drSenderIndex                  = 4
	drRecvrIndex                   = 5
	moSenderIndex                  = 5
	moRecvrIndex                   = 4
	drMsgIndex                     = 24
	moMsgIndex                     = 24
	drSctsIndex                    = 18
	moSctsIndex                    = 18
	ackIndex                       = 4
	xserIndex                      = 34
	errMsgOffset                   = 2
	errCodeOffset                  = 3
	errCodeTimeout                 = "010"
)
