package examples

import "fmt"

// TypeConstruction shows, how type constructions are handled,
// when collecting error codes in the analyser.
//
// Errors:
//
// From Error type construction:
//    - examples-error-not-implemented --
// From Error3 type construction:
//    - examples-error-closed --
//    - examples-error-flagged --
//    - examples-error-unknown --
func TypeConstruction(flag1, flag2 bool) error { // want TypeConstruction:"ErrorCodes: examples-error-closed examples-error-flagged examples-error-not-implemented examples-error-unknown"
	if flag1 {
		return &Error{"examples-error-not-implemented"}
	}

	err := &Error3{false, "examples-error-closed"}
	if flag2 {
		return err
	}
	return nil
}

// AssignmentToErrorCodeField demonstrates,
// how assignments to the error code field of a type are handled,
// when collecting error codes in the analyser.
//
// Errors:
//
//    - examples-error-unkown --
//    - examples-error-zero   --
//    - examples-error-one    --
func AssignmentToErrorCodeField(typ int) error { // want AssignmentToErrorCodeField:"ErrorCodes: examples-error-one examples-error-unkown examples-error-zero"
	err := &Error{"examples-error-unkown"}
	switch typ {
	case 0:
		err.TheCode = "examples-error-zero"
	case 1:
		err.TheCode = "examples-error-one"
	default:
		// Not allowed, because it cannot be statically analysed:
		err.TheCode = fmt.Sprintf("examples-error-%d", typ) // want "error code has to be constant value or error code parameter"
	}
	return err
}

// FunctionCall demonstrates, how function calls are handled,
// when collecting error codes in the analyser.
//
// Errors:
//
//    - examples-error-failed       -- failed to open file
//    - examples-error-invalid-name -- invalid file name
func FunctionCall() error { // want FunctionCall:"ErrorCodes: examples-error-failed examples-error-invalid-name"
	return TryOpen("example.txt")
}

// Errors:
//
//    - examples-error-failed       -- failed to open file
//    - examples-error-invalid-name -- invalid file name
func TryOpen(fileName string) error { // want TryOpen:"ErrorCodes: examples-error-failed examples-error-invalid-name"
	if fileName != "exmaple.txt" {
		return &Error{"examples-error-invalid-name"}
	}
	return &Error{"examples-error-failed"}
}
