package methods

type (
	A string
	B struct{}
	C struct {
		U, V string
	}
	D struct {
		A
		B
		C
	}
	X string
	Y struct{}
	Z struct {
		S int
		T string
	}
)

// Errors:
//
//    - easy-error -- is always returned
func (A) Easy1() error { // want Easy1:"ErrorCodes: easy-error"
	return &Error{"easy-error"}
}

// Errors:
//
//    - easy-error -- is always returned
func (B) Easy1() error { // want Easy1:"ErrorCodes: easy-error"
	return &Error{"easy-error"}
}

// Errors:
//
//    - easy-error -- is always returned
func (C) Easy1() error { // want Easy1:"ErrorCodes: easy-error"
	return &Error{"easy-error"}
}

// Errors:
//
//    - easy-error -- is always returned
func (D) Easy1() error { // want Easy1:"ErrorCodes: easy-error"
	return &Error{"easy-error"}
}

// Errors:
//
//    - easy-error -- is always returned
func (*A) Easy2() error { // want Easy2:"ErrorCodes: easy-error"
	return &Error{"easy-error"}
}

// Errors:
//
//    - easy-error -- is always returned
func (*B) Easy2() error { // want Easy2:"ErrorCodes: easy-error"
	return &Error{"easy-error"}
}

// Errors:
//
//    - easy-error -- is always returned
func (*C) Easy2() error { // want Easy2:"ErrorCodes: easy-error"
	return &Error{"easy-error"}
}

// Errors:
//
//    - easy-error -- is always returned
func (*D) Easy2() error { // want Easy2:"ErrorCodes: easy-error"
	return &Error{"easy-error"}
}

// Unused, here to test that the correct methodA is found.
func (*X) methodA() error {
	return &Error{"x-error"}
}

func (*A) methodA() error {
	return &Error{"a-error"}
}

func (*B) methodB() error {
	return &Error{"b-error"}
}

func (*C) methodC() error {
	return &Error{"c-error"}
}

// Unused, here to test that the correct methodA is found.
func (*Y) methodA() error {
	return &Error{"y-error"}
}

// Errors:
//
//    - a-error --
func (a *A) MethodCall() error { // want MethodCall:"ErrorCodes: a-error"
	return a.methodA()
}

// Errors:
//
//    - a-error --
func (a A) IndirectMethodCall() error { // want IndirectMethodCall:"ErrorCodes: a-error"
	return a.methodA()
}

// Errors:
//
//    - a-error --
//    - b-error --
//    - c-error --
func (d *D) PromotedCall() error { // want PromotedCall:"ErrorCodes: a-error b-error c-error"
	switch {
	case true:
		return d.methodA()
	case true:
		return d.methodB()
	case true:
		return d.methodC()
	}
	return nil
}

// Errors:
//
//    - a-error --
//    - b-error --
//    - c-error --
func (d D) IndirectPromotedCall() error { // want IndirectPromotedCall:"ErrorCodes: a-error b-error c-error"
	switch {
	case true:
		return d.methodA()
	case true:
		return d.methodB()
	case true:
		return d.methodC()
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
