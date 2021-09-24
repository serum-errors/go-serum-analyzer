package inner1

// ExportedFunc1 is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned
func ExportedFunc1() error { // want ExportedFunc1:"ErrorCodes: hello-error"
	return &Error{"hello-error"}
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }

type UnusedError struct { // want UnusedError:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *UnusedError) Code() string  { return e.TheCode }
func (e *UnusedError) Error() string { return e.TheCode }
