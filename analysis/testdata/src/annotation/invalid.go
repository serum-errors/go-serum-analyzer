package annotation

// Errors:
//
//    - some-error -- error forced by annotation
func InvalidOverwriteReturn1() error { // want InvalidOverwriteReturn1:"ErrorCodes: some-error"
	// Error Codes: overwritten-error
	return &Error{"some-error"} // want "error in annotation: expected '=', '\\+=', '-=', '\\+code', or '-code' after 'Error Codes' indicator"
}
