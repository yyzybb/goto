package common

const (
	RpcError_Ok byte = iota
	RpcError_NotConn
	RpcError_PackTooLarge
	RpcError_Overwrite
	RpcError_SendTimeout
	RpcError_RecvTimeout
	RpcError_NoMethod
	RpcError_ParseError
	RpcError_MagicCodeError
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
	return "error:{" + string(e.code) + ", " + e.what + "}"
}
