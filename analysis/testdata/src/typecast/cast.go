package typecast

// Errors:
//
//    - string-error --
func TypeCast() error { // want TypeCast:"ErrorCodes: string-error"
	return StringError("some message about the error")
}
