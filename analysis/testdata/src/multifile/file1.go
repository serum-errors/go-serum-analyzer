package multifile

// Errors:
//
//    - func1-error --
//    - func2-error --
func Func1() error { // want Func1:"ErrorCodes: func1-error func2-error"
	if false {
		return Func2()
	}
	return &Error{"func1-error"}
}

// Errors:
//
//    - string-error --
func DoString() error { // want DoString:"ErrorCodes: string-error"
	return StringError("some message about the error")
}

// Errors:
//
//    - local-error --
func GlobalFromOtherFile() error { // want GlobalFromOtherFile:"ErrorCodes: local-error"
	if true {
		return globalError // want "returned error may not be a parameter, receiver or global variable"
	}
	return &Error{"local-error"}
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
