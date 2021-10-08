package multifile

var globalError error

// Errors:
//
//    - func2-error --
func Func2() error { // want Func2:"ErrorCodes: func2-error"
	return &Error{"func2-error"}
}

type StringError string // want StringError:`ErrorType{Field:<nil>, Codes:string-error}`

func (StringError) Code() string  { return "string-error" }
func (StringError) Error() string { return "StringError" }
