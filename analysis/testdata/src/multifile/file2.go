package multifile

var globalError error

// Errors:
//
//    - func2-error --
func Func2() error { // want Func2:"ErrorCodes: func2-error"
	return &Error{"func2-error"}
}
