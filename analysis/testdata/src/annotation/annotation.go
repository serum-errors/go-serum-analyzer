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
//    - overwritten-3-error --
//    - overwritten-4-error --
func OverwriteReturn() error { // want OverwriteReturn:"ErrorCodes: overwritten-1-error overwritten-2-error overwritten-3-error overwritten-4-error"
	switch {
	case true:
		// Error Codes = overwritten-1-error
		return &Error{"some-1-error"}
	case true:
		err := &Error{}
		err.TheCode = "some-2-error"
		// Error Codes=overwritten-2-error
		return err
	case true:
		// Error Codes = overwritten-3-error, overwritten-4-error, overwritten-3-error
		return nil
	}
	return nil
}

// Errors: none
func OverwriteReturn3() error { // want OverwriteReturn3:"ErrorCodes:"
	// Error Codes =
	return &Error{"some-error"}
}

// Errors:
//
//    - some-error        -- found by analysis
//    - overwritten-error -- forced by annotation
func AddReturn1() error { // want AddReturn1:"ErrorCodes: overwritten-error some-error"
	// Error Codes += overwritten-error
	return &Error{"some-error"}
}

// Errors:
//
//    - some-1-error        --
//    - some-2-error        --
//    - overwritten-1-error --
//    - overwritten-2-error --
//    - overwritten-3-error --
//    - overwritten-4-error --
func AddReturn2() error { // want AddReturn2:"ErrorCodes: overwritten-1-error overwritten-2-error overwritten-3-error overwritten-4-error some-1-error some-2-error"
	switch {
	case true:
		// Error Codes += overwritten-1-error
		return &Error{"some-1-error"}
	case true:
		err := &Error{}
		err.TheCode = "some-2-error"
		// Error Codes+=overwritten-2-error
		return err
	case true:
		// Error Codes += overwritten-3-error, overwritten-4-error, overwritten-3-error
		return nil
	}
	return nil
}

// Errors: none
func SubReturn1() error { // want SubReturn1:"ErrorCodes:"
	// Error Codes -= some-error
	return &Error{"some-error"}
}

// Errors:
//
//    - assigned-error      --
//    - some-1-error        --
//    - overwritten-2-error --
//    - overwritten-3-error --
//    - overwritten-4-error --
func SubReturn2() error { // want SubReturn2:"ErrorCodes: assigned-error overwritten-2-error overwritten-3-error overwritten-4-error some-1-error"
	switch {
	case true:
		// Error Codes -= overwritten-1-error, some-2-error, some-2-error
		return AddReturn2()
	case true:
		err := &Error{}
		err.TheCode = "assigned-error"
		// Error Codes-=assigned-error
		return err // With how it's currently implemented it would be difficult to remove assigned codes. Use overwrite in this case (see below).
	case true:
		err := &Error{}
		err.TheCode = "unkown-error"
		// Error Codes =
		return err
	case true:
		// Error Codes -= overwritten-3-error, overwritten-4-error, overwritten-3-error
		return nil
	}
	return nil
}

// Errors:
//
//    - overwritten-error --
func AddSubReturn1() error { // want AddSubReturn1:"ErrorCodes: overwritten-error"
	// Error Codes -some-error +overwritten-error
	return &Error{"some-error"}
}

// Errors:
//
//    - added-1-error       --
//    - added-2-error       --
//    - assigned-error      --
//    - some-1-error        --
//    - overwritten-2-error --
//    - overwritten-3-error --
//    - overwritten-4-error --
func AddSubReturn2() error { // want AddSubReturn2:"ErrorCodes: added-1-error added-2-error assigned-error overwritten-2-error overwritten-3-error overwritten-4-error some-1-error"
	switch {
	case true:
		// Error Codes -overwritten-1-error -some-2-error +added-1-error -some-2-error
		return AddReturn2()
	case true:
		err := &Error{}
		err.TheCode = "assigned-error"
		// Error Codes -assigned-error
		return err // With how it's currently implemented it would be difficult to remove assigned codes.
	case true:
		// Error Codes -overwritten-3-error -overwritten-4-error -overwritten-3-error
		return nil
	case true:
		// Error Codes -overwritten-3-error    +added-2-error    -overwritten-3-error
		return nil
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
