package annotation

// Errors:
//
//    - some-error -- error forced by annotation
func InvalidOverwriteReturn1() error { // want InvalidOverwriteReturn1:"ErrorCodes: some-error"
	// Error Codes: overwritten-error
	return &Error{"some-error"} // want "error in annotation: expected '=', '\\+=', '-=', '\\+code', or '-code' after 'Error Codes' indicator"
}

// Errors: none
func InvalidOverwriteReturn2() error { // want InvalidOverwriteReturn2:"ErrorCodes:"
	switch {
	case true:
		// Error Codes = -overwritten-error
		return nil // want "invalid error code in annotation: should match (.*)"
	case true:
		// Error Codes overwritten-error
		return nil // want "error in annotation: expected '=', '\\+=', '-=', '\\+code', or '-code' after 'Error Codes' indicator"
	case true:
		// Error Codes = overwritten-error,,other-error
		return nil // want "invalid error code in annotation: should match (.*)"
	case true:
		// Error Codes = overwritten-error,
		return nil // want "invalid error code in annotation: should match (.*)"
	}
	return nil
}

// Errors: none
func InvalidAddReturn() error { // want InvalidAddReturn:"ErrorCodes:"
	switch {
	case true:
		// Error Codes += +overwritten-error
		return nil // want "invalid error code in annotation: should match (.*)"
	case true:
		// Error Codes += overwritten-error,,other-error
		return nil // want "invalid error code in annotation: should match (.*)"
	case true:
		// Error Codes += overwritten-error,
		return nil // want "invalid error code in annotation: should match (.*)"
	}
	return nil
}

// Errors: none
func InvalidSubReturn() error { // want InvalidSubReturn:"ErrorCodes:"
	switch {
	case true:
		// Error Codes -= +overwritten-error
		return nil // want "invalid error code in annotation: should match (.*)"
	case true:
		// Error Codes -= overwritten-error,,other-error
		return nil // want "invalid error code in annotation: should match (.*)"
	case true:
		// Error Codes -= overwritten-error,
		return nil // want "invalid error code in annotation: should match (.*)"
	}
	return nil
}

// Errors: none
func InvalidAddSubReturn() error { // want InvalidAddSubReturn:"ErrorCodes:"
	switch {
	case true:
		// Error Codes ++overwritten-error
		return nil // want "invalid error code in annotation: should match (.*)"
	case true:
		// Error Codes -overwritten-error - -other-error
		return nil // want "invalid error code in annotation: should match (.*)"
	case true:
		// Error Codes -overwritten-error other-error
		return nil // want "invalid error code in annotation: code has to start with '\\+' or '-'"
	}
	return nil
}
