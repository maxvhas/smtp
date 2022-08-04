package smtp

type ResponseCode int

const (
	RespFAILURE ResponseCode = 554
	RespOK                   = 220
)
