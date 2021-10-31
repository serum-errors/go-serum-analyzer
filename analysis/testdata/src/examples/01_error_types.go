package examples

import (
	"fmt"
	"strings"
)

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Error() string { return e.TheCode }
func (e *Error) Code() string  { return e.TheCode }

type Error2 string // want Error2:`ErrorType{Field:<nil>, Codes:examples-error-disconnect examples-error-empty examples-error-unknown}`

const (
	errorPrefix     = "examples-error-"
	ErrorEmpty      = errorPrefix + "empty"
	ErrorDisconnect = errorPrefix + "disconnect"
	ErrorUnknown    = errorPrefix + "unknown"
)

func (e Error2) Error() string { return fmt.Sprintf("%s: %s", e.Code(), e) }
func (e Error2) Code() string {
	switch {
	case e == "":
		return ErrorEmpty
	case strings.HasPrefix(string(e), "peer disconnected"):
		return ErrorDisconnect
	default:
		return ErrorUnknown
	}
}

// Errors:
//
//    - examples-error-empty      --
//    - examples-error-disconnect --
//    - examples-error-unknown    --
func ReturnError2() Error2 { // want ReturnError2:"ErrorCodes: examples-error-disconnect examples-error-empty examples-error-unknown"
	return Error2("error message")
}

type Error3 struct { // want Error3:`ErrorType{Field:{Name:"code", Position:1}, Codes:examples-error-flagged examples-error-unknown}`
	flag bool
	code string
}

func (e *Error3) Error() string { return e.code }
func (e *Error3) Code() string {
	if e.flag {
		e.code = "examples-error-flagged"
	}
	if e.code == "" {
		return "examples-error-unknown"
	}
	return e.code
}

// Errors:
//
//    - examples-error-flagged         --
//    - examples-error-unknown         --
//    - examples-error-not-implemented --
func ReturnError3() *Error3 { // want ReturnError3:"ErrorCodes: examples-error-flagged examples-error-not-implemented examples-error-unknown"
	return &Error3{false, "examples-error-not-implemented"}
}
