package typecast

type StringError string // want StringError:`ErrorType{Field:<nil>, Codes:string-error}`

func (StringError) Code() string      { return "string-error" }
func (StringError) Error() string     { return "StringError" }
func (s StringError) Message() string { return string(s) }
