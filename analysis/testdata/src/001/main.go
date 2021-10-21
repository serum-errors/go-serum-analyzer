package main

import (
	"fmt"
)

func main() {
	One()
}

/*
	One is a demo function.
	It calls another function that returns an error.

	Errors:

		- hello-error -- always returned.
*/
func One() error { // want One:"ErrorCodes: hello-error"
	return Two()
}

// Two is a demo function.
// It returns an error value it creates itself.
//
// Errors:
//
//    - hello-error -- is always returned.
func Two() error { // want Two:"ErrorCodes: hello-error"
	return &Error{"hello-error"}
}

func Three() *Error { // want `function "Three" is exported, but does not declare any error codes`
	return &Error{"hello-error-literal"}
}

// Four is a demo function with multiple returns.
//
// Errors:
//
//    - zonk-error -- is always returned.
func Four() (string, error) { // want Four:"ErrorCodes: zonk-error"
	return "something", &Error{"zonk-error"}
}

// Five is a demo function with multiple returns
// which is returning the result of another function
// (which *also* has multiple returns).
//
// Errors:
//
//    - zonk-error -- is always returned.
func Five() (interface{}, error) { // want Five:"ErrorCodes: zonk-error"
	return Four()
}

// Five is a demo function with multiple returns
// which is returning the result of another function
// (which *also* has multiple returns).
//
// Errors:
//
//    - hello-error -- is sometimes returned.
//    - zonk-error -- is returned at other times.
func Six(flip bool) error { // want Six:"ErrorCodes: hello-error zonk-error"
	err := (Two)()
	if flip {
		_, err = Four()
	}
	return err
}

// Seven is ... probably out of scope.
//
// Errors:
//
//    - hello-error -- is always returned.
func Seven() error { // want Seven:"ErrorCodes: hello-error"
	uff := &Error{}
	uff.TheCode = "hello-error"
	var err error
	err = uff
	return err
}

// Eight isn't going to fulfill what it says it will do.
// It also makes a call to another package.
//
// Errors:
//
//    - hello-error -- is a lie, won't actually happen.
func Eight() error { // want Eight:"ErrorCodes: hello-error" `function "Eight" has a mismatch of declared and actual error codes: unused codes: \[hello-error]`
	return fmt.Errorf("not a nice structural error") // want `function "Errorf" in package "fmt" does not declare error codes`
}

// Named returns an error by named return arguments.
//
// Errors:
//
//    - named-error -- is always returned
func Named() (something string, err error) { // want Named:"ErrorCodes: named-error"
	something = "value"
	err = &Error{"named-error"}
	return
}

// Shadowed returns an error by named return arguments.
//
// Errors:
//
//    - named-error -- is always returned
func Shadowed() (err error) { // want Shadowed:"ErrorCodes: named-error"
	err = &Error{"named-error"}
	{
		var err string // shadowed
		err = "some value"
		_ = err
	}
	return
}

// IdentLoop does assignments in an attempt to make the analysis fall into an endless loop.
//
// Errors:
//
//    - some-error -- is always returned
func IdentLoop() error { // want IdentLoop:"ErrorCodes: some-error"
	err1 := &Error{"some-error"}
	err2 := err1
	err3 := err2
	err1 = err3
	return err1
}

// PointerOfLocal returns a pointer to a locally created error.
//
// Errors:
//
//    - value-error -- is always returned
func PointerOfLocal() error { // want PointerOfLocal:"ErrorCodes: value-error"
	err := Error{"value-error"}
	if false {
		err2 := &err
		return err2
	}
	return &err
}

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

type Io interface { // want Io:"ErrorInterface: FetchFromUrl"
	// Errors:
	//
	//    - io-error -- returned if transfer fails
	FetchFromUrl(url string) error // want FetchFromUrl:"ErrorCodes: io-error"
}

type IoImpl struct{}

// Errors:
//
//    - io-error -- returned if transfer fails
func (IoImpl) FetchFromUrl(url string) error { // want FetchFromUrl:"ErrorCodes: io-error"
	failed := false
	// do some fancy io stuff ...
	if failed {
		return &Error{"io-error"}
	}
	return nil
}

type IoMock struct{}

// Errors: none -- this method only returns error to comply with the IO interface.
func (IoMock) FetchFromUrl(url string) error { // want FetchFromUrl:"ErrorCodes:"
	// this is only a mock and does not return errors
	return nil
}

func ConsumingIoImplementations() {
	var io1 Io = IoImpl{}
	var io2 Io = IoMock{}
	_, _ = io1, io2
}

// Errors:
//
//    - var-error --
func DefinitionWithoutAssignment() error { // want DefinitionWithoutAssignment:"ErrorCodes: var-error"
	var err error = &Error{"var-error"}
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
