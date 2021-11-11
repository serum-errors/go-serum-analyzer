package annotation

// Errors:
//
//    - overwritten-error -- error forced by annotation
func OverwriteReturn() error {
	// Error Codes = overwritten-error
	return &Error{"some-error"}
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
