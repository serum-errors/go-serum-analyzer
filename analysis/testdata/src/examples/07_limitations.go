package examples

// Errors:
//
//    - examples-error-invalid --
func CallModifyError() error { // want CallModifyError:"ErrorCodes: examples-error-invalid"
	err := &Error{"examples-error-invalid"}
	ModifyError(err)
	return err
}

func ModifyError(err *Error) {
	err.TheCode = "some invalid value"
}

// Errors:
//
//    - example-error-unreachable -- is never actually returned
func DeadBranchError() error { // want DeadBranchError:"ErrorCodes: example-error-unreachable"
	if false {
		return &Error{"example-error-unreachable"}
	}
	return nil
}

func ErrorNotLast() (error, string) { // want "error should be returned as the last argument"
	return nil, ""
}
