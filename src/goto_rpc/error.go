package goto_rpc

const (
	RpcError_Ok byte = iota
	RpcError_NotEstab
	RpcError_PackTooLarge
	RpcError_Overwrite
	RpcError_SendTimeout
	RpcError_RecvTimeout
	RpcError_NoMethod
	RpcError_ParseError
	RpcError_MagicCodeError
	RpcError_RepeatMethod
)

type IError interface {
	Code() int
	Error() string
}

type Error struct {
	code int
	what string
}

func NewError(enum byte) *Error {
	return &Error{int(enum), ""}
}

func NewErrorMsg(code int, what string) *Error {
	return &Error{code, what}
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Error() string {
	return "error:{" + e.CodeToString() + ", " + e.what + "}"
}

func (e *Error) CodeToString() string {
	switch byte(e.code) {
		case RpcError_Ok: return "RpcError_Ok"
		case RpcError_NotEstab: return "RpcError_NotEstab"
		case RpcError_PackTooLarge: return "RpcError_PackTooLarge"
		case RpcError_Overwrite: return "RpcError_Overwrite"
		case RpcError_SendTimeout: return "RpcError_SendTimeout"
		case RpcError_RecvTimeout: return "RpcError_RecvTimeout"
		case RpcError_NoMethod: return "RpcError_NoMethod"
		case RpcError_ParseError: return "RpcError_ParseError"
		case RpcError_MagicCodeError: return "RpcError_MagicCodeError"
		case RpcError_RepeatMethod: return "RpcError_RepeatMethod"
	default:
		return "unkown"
    }
}
