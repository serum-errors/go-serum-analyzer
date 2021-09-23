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

// Unused, here to test that the correct valueMethodA is found.
func (X) valueMethodA() error {
	return &Error{"x-value-error"}
}

func (*A) methodA() error {
	return &Error{"a-error"}
}

func (A) valueMethodA() error {
	return &Error{"a-value-error"}
}

func (*B) methodB() error {
	return &Error{"b-error"}
}

func (B) valueMethodB() error {
	return &Error{"b-value-error"}
}

func (*C) methodC() error {
	return &Error{"c-error"}
}

func (C) valueMethodC() error {
	return &Error{"c-value-error"}
}

// Unused, here to test that the correct methodA is found.
func (*Y) methodA() error {
	return &Error{"y-error"}
}

// Unused, here to test that the correct valueMethodA is found.
func (Y) valueMethodA() error {
	return &Error{"y-value-error"}
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
//    - a-value-error --
func (a *A) DereferencedMethodCall() error { // want DereferencedMethodCall:"ErrorCodes: a-value-error"
	return a.valueMethodA()
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

// Errors:
//
//    - a-value-error --
//    - b-value-error --
//    - c-value-error --
func (d *D) DereferencedPromotedCall() error { // want DereferencedPromotedCall:"ErrorCodes: a-value-error b-value-error c-value-error"
	switch {
	case true:
		return d.valueMethodA()
	case true:
		return d.valueMethodB()
	case true:
		return d.valueMethodC()

	}
	return nil
}

// Errors:
//
//    - a-error --
//    - b-error --
//    - c-error --
//    - x-error --
func FunctionCallingMethods(a *A, b B, d *D, x *X) error { // want FunctionCallingMethods:"ErrorCodes: a-error b-error c-error x-error"
	switch {
	case true:
		return a.methodA()
	case true:
		return b.methodB()
	case true:
		return d.methodC()
	case true:
		return x.methodA()
	}
	return nil
}

// Errors:
//
//    - a-error --
//    - b-error --
//    - c-error --
func ArrayMethodCall(ds []D) error { // want ArrayMethodCall:"ErrorCodes: a-error b-error c-error"
	switch {
	case true:
		return ds[0].methodA()
	case true:
		return ds[1].methodB()
	case true:
		for _, d := range ds {
			return d.methodC()
		}
	}
	return nil
}

// Errors:
//
//    - a-error --
func Struct(s *struct{ D D }) error { // want Struct:"ErrorCodes: a-error"
	return s.D.methodA()
}

// Errors:
//
//    - c-error --
func (d *D) MultipleResults() (*D, A, error) { // want MultipleResults:"ErrorCodes: c-error"
	return d, d.A, d.methodC()
}

// Errors:
//
//    - a-error --
//    - c-error --
func CallMultipleResults(d *D) error { // want CallMultipleResults:"ErrorCodes: a-error c-error"
	_, a, err := d.MultipleResults()
	if err != nil {
		return err
	}

	return a.methodA()
}

func (d *D) self() *D {
	return d
}

// Errors:
//
//    - a-error --
func (d *D) Chained() error { // want Chained:"ErrorCodes: a-error"
	return d.self().self().self().self().methodA()
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
