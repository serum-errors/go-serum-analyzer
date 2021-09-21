package main

type (
	A string
	B struct{}
	C struct {
		X, Y string
	}
	D struct {
		A
		B
		C
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

func (*A) methodA() error {
	return &Error{"a-error"}
}

func (*B) methodB() error {
	return &Error{"b-error"}
}

func (*C) methodC() error {
	return &Error{"c-error"}
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
