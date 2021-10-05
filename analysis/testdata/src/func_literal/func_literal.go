package funcliteral

// Errors:
//
//    - some-error --
func AssignmentInLambda() *Error { // want AssignmentInLambda:"ErrorCodes: some-error"
	var err *Error
	func() {
		err = &Error{"some-error"}
	}()
	func() {
		err := &Error{"no-error"} // shadowed!
		_ = err
	}()
	return err
}

// Errors:
//
//    - some-error --
func OutOfBounds() *Error { // want OutOfBounds:"ErrorCodes: some-error"
	switch {
	case true:
		return &Error{"some-error"}
	case true:
		return func() *Error { // want "unnamed functions are not supported in error code analysis"
			return &Error{"other-error"}
		}()
	case true:
		return &Error{func() string { return "other-error" }()} // want "error code field has to be instantiated by constant value"
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
