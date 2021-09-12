package main

// UnusedCode declares an error which is never returned.
//
// Errors:
//
//    - unused-error -- never returned
func UnusedCode() error { /*
		want
			UnusedCode:"ErrorCodes: unused-error"
			`function "UnusedCode" has a mismatch of declared and actual error codes: unused codes: \[unused-error]` */
	return nil
}

// MissingCode returns an error which is not declared.
//
// Errors:
//
//    - hello-error -- error that might be returned
func MissingCode() error { /*
		want
			MissingCode:"ErrorCodes: hello-error"
			`function "MissingCode" has a mismatch of declared and actual error codes: missing codes: \[missing-error]` */
	if false {
		return &Error{"hello-error"}
	}
	return &Error{"missing-error"}
}

// MissingAndUnusedCode returns an error which is not declared,
// and declares an error which is never returned.
//
// Errors:
//
//    - unused-error -- never returned
func MissingAndUnusedCode() error { /*
		want
			MissingAndUnusedCode:"ErrorCodes: unused-error"
			`function "MissingAndUnusedCode" has a mismatch of declared and actual error codes: missing codes: \[missing-error] unused codes: \[unused-error]` */
	return &Error{"missing-error"}
}

// MultipleMismatchedCodes has multiple error code mismatches.
//
// Errors:
//
//    - hello-error     -- error that might be returned
//    - unused-error-ab -- never returned
//    - unused-error-aa -- never returned
func MultipleMismatchedCodes() error { /*
		want
			MultipleMismatchedCodes:"ErrorCodes: hello-error, unused-error-aa, unused-error-ab"
			`function "MultipleMismatchedCodes" has a mismatch of declared and actual error codes: missing codes: \[missing-error-a missing-error-ab missing-error-cc] unused codes: \[unused-error-aa unused-error-ab]` */
	switch {
	case true:
		return &Error{"missing-error-ab"}
	case true:
		return &Error{"missing-error-a"}
	case true:
		return &Error{"missing-error-cc"}
	default:
		return &Error{"hello-error"}
	}
}
