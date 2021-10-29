package examples

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Error() string { return e.TheCode }
func (e *Error) Code() string  { return e.TheCode }
