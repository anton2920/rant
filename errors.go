package main

import "unsafe"

type E struct {
	Message string
	Code    int
}

const (
	/* From <errno.h>. */
	ENOENT      = 2
	EINTR       = 4
	EPIPE       = 32
	EAGAIN      = 35
	EWOULDBLOCK = EAGAIN
	ECONNRESET  = 54
	ETIMEDOUT   = 60
)

func (e E) Error() string {
	var buffer [512]byte

	n := copy(buffer[:], e.Message)
	if e.Code < 0 {
		e.Code = -e.Code
	}
	n += SlicePutPositiveInt(buffer[n:], e.Code)

	return string(unsafe.Slice(&buffer[0], n))
}

func NewError(msg string, code int) error {
	return error(E{Message: msg, Code: code})
}

func Fatal(msg string, code int) {
	if code < 0 {
		code = -code
	}
	println(msg, code)
	Exit(1)
}

func FatalError(err error) {
	println(err.Error())
	Exit(1)
}
