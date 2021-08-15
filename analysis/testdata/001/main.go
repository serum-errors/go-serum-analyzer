package main

func main() {
	One()
}

/*
	One is a demo function.

	Errors:

		- hello-error -- always returned.
*/
func One() error {
	return Two()
}

// Two is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned.
func Two() error {
	return &Error{"hello-error"}
}

func Three() *Error {
	return &Error{"hello-error-literal"}
}

type Error struct {
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
