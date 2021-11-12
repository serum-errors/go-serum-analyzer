package annotation

// Errors:
//
//    - overwritten-error -- error forced by annotation
func OverwriteReturn1() error { // want OverwriteReturn1:"ErrorCodes: overwritten-error"
	// Error Codes = overwritten-error
	return &Error{"some-error"}
}

// Errors:
//
//    - overwritten-1-error --
//    - overwritten-2-error --
func OverwriteReturn() error { // want OverwriteReturn:"ErrorCodes: overwritten-1-error overwritten-2-error"
	switch {
	case true:
		// Error Codes = overwritten-1-error
		return &Error{"some-1-error"}
	case true:
		err := &Error{}
		err.TheCode = "some-2-error"
		// Error Codes=overwritten-2-error
		return err
	}
	return nil
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
