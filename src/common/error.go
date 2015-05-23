package common

const (
	RpcError_NotConn = iota
	RpcError_PackTooLarge
	RpcError_Overwrite
	RpcError_SendTimeout
	RpcError_RecvTimeout
	RpcError_NoService
	RpcError_NoMethod
	RpcError_Fatal
)

type Error struct {
	code int
	what string
}

func NewError(code int, what string) *Error {
	return &Error{code, what}
}

func (e *Error) Error() string {
	return "error:{" + string(e.code) + ", " + e.what + "}"
}
