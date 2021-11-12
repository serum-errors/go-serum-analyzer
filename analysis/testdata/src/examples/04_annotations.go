package examples

// Overwrite uses an annotation to overwrite the actual error code examples-error-unknown
// with examples-error-overwritten for the error code analysis.
//
// Errors:
//
//    - examples-error-overwritten -- error forced by annotation
func Overwrite() error { // want Overwrite:"ErrorCodes: examples-error-overwritten"
	// Error Codes = examples-error-overwritten
	return &Error{"examples-error-unknown"}
}

// OverwriteMultiple demonstrates how commas can be used in the overwrite annotation
// to set multiple error codes.
//
// Errors:
//
//    - examples-error-one -- error forced by annotation
//    - examples-error-two -- error forced by annotation
func OverwriteMultiple(index int) error { // want OverwriteMultiple:"ErrorCodes: examples-error-one examples-error-two"
	errors := []error{
		nil,
		&Error{"examples-error-one"},
		&Error{"examples-error-two"},
	}
	// Error Codes = examples-error-one, examples-error-two
	return errors[index]
}

// Errors:
//
//    - examples-error-one   --
//    - examples-error-two   --
//    - examples-error-three --
func MultipleCodes() *Error { // want MultipleCodes:"ErrorCodes: examples-error-one examples-error-three examples-error-two"
	switch {
	case true:
		return &Error{"examples-error-one"}
	case true:
		return &Error{"examples-error-two"}
	case true:
		return &Error{"examples-error-three"}
	}
	return nil
}

func HandleErrorOne(err *Error) { /* ... */ }

// SubtractCode handles the error code examples-error-one and therefore uses an
// annotation to get rid of it when returning the error.
//
// Errors:
//
//    - examples-error-two   --
//    - examples-error-three --
func SubtractCode() error { // want SubtractCode:"ErrorCodes: examples-error-three examples-error-two"
	switch err := MultipleCodes(); err.Code() {
	case "": // Code could always return an empty string.
		return nil
	case "examples-error-one":
		HandleErrorOne(err)
		return nil
	default:
		// Error Codes -= examples-error-one
		return err
	}
}

// Errors:
//
//    - examples-error-one   --
//    - examples-error-extra --
func AddSubCode() error { // want AddSubCode:"ErrorCodes: examples-error-extra examples-error-one"
	err := MultipleCodes()
	// Error Codes -examples-error-two -examples-error-three +examples-error-extra
	return err
}

// AssignmentProblem showcases error code field assignment not being
// removed when using the remove annotation.
//
// Errors:
//
//    - assigned-error --
func AssignmentProblem() error { // want AssignmentProblem:"ErrorCodes: assigned-error"
	err := &Error{}
	err.TheCode = "assigned-error"
	// Error Codes -= assigned-error
	return err
}
