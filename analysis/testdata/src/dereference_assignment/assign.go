package dereferenceassignment

// Errors:
//
//    - some-error --
//    - other-error --
func DereferenceAssignment() error { // want DereferenceAssignment:"ErrorCodes: other-error some-error"
	err := Error{"some-error"}
	err2 := &err
	*err2 = Error{"other-error"}
	return &err
}

// Errors:
//
//    - some-error --
//    - other-error --
func DereferenceAssignment2() error { // want DereferenceAssignment2:"ErrorCodes: other-error some-error"
	err := &Error{"some-error"}
	*err = Error{"other-error"}
	return err
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
