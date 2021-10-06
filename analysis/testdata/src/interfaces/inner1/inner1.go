package inner1

type ImplementOuter1 struct{}

// Errors:
//
//    - interface-1-error --
func (ImplementOuter1) InterfaceMethod1() error { // want InterfaceMethod1:"ErrorCodes: interface-1-error"
	return &Error{"interface-1-error"}
}

// Errors:
//
//    - interface-3-error --
func (ImplementOuter1) InterfaceMethod2(a, b string) error { // want InterfaceMethod2:"ErrorCodes: interface-3-error"
	return &Error{"interface-3-error"}
}

func (ImplementOuter1) MethodWithoutError(a, b string) string {
	return a
}

type ImplementOuterPointer1 struct{}

// Errors:
//
//    - interface-2-error --
func (*ImplementOuterPointer1) InterfaceMethod1() error { // want InterfaceMethod1:"ErrorCodes: interface-2-error"
	return &Error{"interface-2-error"}
}

// Errors:
//
//    - interface-4-error --
func (*ImplementOuterPointer1) InterfaceMethod2(a, b string) error { // want InterfaceMethod2:"ErrorCodes: interface-4-error"
	return &Error{"interface-4-error"}
}

func (*ImplementOuterPointer1) MethodWithoutError(a, b string) string {
	return a
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
