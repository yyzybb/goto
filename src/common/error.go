package common

type Error struct {
	what string
	code int
}

func (e *Error) Error() string {
	return "error:{" + string(e.code) + ", " + e.what + "}"
}
