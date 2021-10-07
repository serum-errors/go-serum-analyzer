package inner1

type Inner1Interface1 interface {
	// InterfaceMethod1 is a method returning an error with error codes declared in doc.
	//
	// Errors:
	//
	//    - interface-1-error -- could potentially be returned
	//    - interface-2-error --
	Inner1Method1() error // want Inner1Method1:"ErrorCodes: interface-1-error interface-2-error"

	// Errors:
	//
	//    - interface-3-error -- could potentially be returned
	//    - interface-4-error --
	Inner1Method2(a, b string) *Error // want Inner1Method2:"ErrorCodes: interface-3-error interface-4-error"

	// MethodWithoutError is just some method that does not return an error,
	// to test if it is correctly ignored in the analysis.
	Inner1MethodWithoutError(a, b string) string
}

type Inner1Interface2 interface {
	Inner1CodeNotDeclared() error // want `function "Inner1CodeNotDeclared" is exported, but does not declare any error codes`
}

type Inner1Interface3 interface {
	Inner1NoCodes() error // want `function "Inner1NoCodes" is exported, but does not declare any error codes`

	// Errors:
	//
	//    - interface-1-error -- could potentially be returned
	//    - interface-2-error --
	Inner1YesCodes() error // want Inner1YesCodes:"ErrorCodes: interface-1-error interface-2-error"
}

// Errors:
//
//    - interface-1-error --
//    - interface-2-error --
//    - interface-3-error --
//    - interface-4-error --
func FunctionForInterface1(a, b string, v Inner1Interface1) error { // want FunctionForInterface1:"ErrorCodes: interface-1-error interface-2-error interface-3-error interface-4-error"
	if false {
		return v.Inner1Method1()
	}
	a = v.Inner1MethodWithoutError(b, a)
	return v.Inner1Method2(a, b)
}

// Errors:
//
//    - some-error --
func FunctionForInterface2(v Inner1Interface2) error { // want FunctionForInterface2:"ErrorCodes: some-error"
	if true {
		return v.Inner1CodeNotDeclared() // want "called function does not declare error codes"
	}
	return &Error{"some-error"}
}

// Errors:
//
//    - interface-1-error --
//    - interface-2-error --
func FunctionForInterface3(v Inner1Interface3) error { // want FunctionForInterface3:"ErrorCodes: interface-1-error interface-2-error"
	if false {
		return v.Inner1NoCodes() // want "called function does not declare error codes"
	}
	return v.Inner1YesCodes()
}

// Errors:
//
//    - interface-1-error --
//    - interface-2-error --
//    - interface-3-error --
//    - interface-4-error --
func FunctionForAllInterfaces(v1 Inner1Interface1, v2 Inner1Interface2, v3 Inner1Interface3) error { // want FunctionForAllInterfaces:"ErrorCodes: interface-1-error interface-2-error interface-3-error interface-4-error"
	switch {
	case true:
		return v2.Inner1CodeNotDeclared() // want "called function does not declare error codes"
	case true:
		return v3.Inner1NoCodes() // want "called function does not declare error codes"
	case true:
		return v3.Inner1YesCodes()
	}
	return v1.Inner1Method2("a", "b")
}

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
