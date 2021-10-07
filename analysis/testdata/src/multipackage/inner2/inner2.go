package inner2

// ExportedFunc2 is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned
func ExportedFunc2() error {
	return &Error{"hello-error"}
}

func CodeNotDeclared() error {
	return &Error{"some-error"}
}

type SomeType struct{}

func (SomeType) CodeNotDeclared() error {
	return &Error{"some-error"}
}

type Error struct {
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }

type UnusedError struct {
	TheCode string
}

func (e *UnusedError) Code() string  { return e.TheCode }
func (e *UnusedError) Error() string { return e.TheCode }
