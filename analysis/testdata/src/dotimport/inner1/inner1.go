package inner1

// ExportedFunc1 is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned
func ExportedFunc1() error { // want ExportedFunc1:"ErrorCodes: hello-error"
	return &Inner1Error{"hello-error"}
}

func CodeNotDeclared1() error { // want `function "CodeNotDeclared1" is exported, but does not declare any error codes`
	return &Inner1Error{"some-error"}
}

type SomeType1 struct{}

func (SomeType1) CodeNotDeclared() error { // want `function "CodeNotDeclared" is exported, but does not declare any error codes`
	return &Inner1Error{"some-error"}
}

type Inner1Error struct { // want Inner1Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Inner1Error) Code() string               { return e.TheCode }
func (e *Inner1Error) Message() string            { return e.TheCode }
func (e *Inner1Error) Details() map[string]string { return nil }
func (e *Inner1Error) Cause() error               { return nil }
func (e *Inner1Error) Error() string              { return e.Message() }

type Inner1UnusedError struct { // want Inner1UnusedError:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Inner1UnusedError) Code() string  { return e.TheCode }
func (e *Inner1UnusedError) Error() string { return e.TheCode }
