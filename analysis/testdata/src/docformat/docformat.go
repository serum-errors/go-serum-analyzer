package docformat

// Correct is a demo function.
// The following is a valid example of an errors docstring.
//
// Errors:
//
//    - hello-error       -- is always returned.
// The following error codes should not occur:
//    - hello-unreachable -- should never be returned.
//    - hello-unreachable --
//
// After a blank line comments in any format may follow.
// Additional blocks starting with 'Errors:' are disallowed.
func Correct() error { // want Correct:"ErrorCodes: hello-error hello-unreachable"
	if false {
		return &Error{"hello-unreachable"}
	}
	return &Error{"hello-error"}
}

// NoError is a demo function.
// Functions that don't return an error, don't have to declare them in an errors docstring.
func NoError() {}

func One() error { // want `function "One" is exported, but does not declare any error codes`
	return one()
}

// OneWithComment is demo function.
// Erros are not documented in this function which should be detected by our analyzer.
func OneWithComment() error { // want `function "OneWithComment" is exported, but does not declare any error codes`
	return one()
}

func one() error { // no problem if the function is not exported
	return nil
}

// Two is a test function.
// It returns an error value it creates itself.
// The following errors docstring is missing a blank line after 'Errors:'.
//
// Errors:
//    - hello-error -- is always returned.
func Two() error { // want `function "Two" has odd docstring: need a blank line after the 'Errors:' block indicator`
	return &Error{"hello-error"}
}

// Three is a demo function.
// The following errors docstring has multiple 'Errors:' block indicators which is invalid.
//
// Errors:
//
// Errors:
//
//    - hello-error -- is always returned.
func Three() error { // want `function "Three" has odd docstring: repeated 'Errors:' block indicator`
	return &Error{"hello-error"}
}

// Four is a demo function.
// The following errors docstring has multiple 'Errors:' block indicators which is invalid.
//
// Errors:
//
//    - hello-error -- is always returned.
//
// Errors:
//
//    - hello-error -- is always returned.
func Four() error { // want `function "Four" has odd docstring: repeated 'Errors:' block indicator`
	return &Error{"hello-error"}
}

// Five is a demo function.
// The following errors docstring has an error code line with an invalid format.
//
// Errors:
//
//    - hello-error - is always returned.
func Five() error { // want `function "Five" has odd docstring: mid block, a line leading with '- ' didnt contain a '--' to mark the end of the code name`
	return &Error{"hello-error"}
}

// Six is a demo function.
// The following errors docstring has an invalid whitespace error code.
//
// Errors:
//
//    - hello-error -- is always returned.
//    - -- is invalid.
func Six() error { // want `function "Six" has odd docstring: an error code can't be purely whitespace`
	return &Error{"hello-error"}
}

// Seven is a demo function.
// The following errors docstring has an invalid whitespace error code.
//
// Errors:
//
//    - hello-error -- is always returned.
//    -             -- is invalid.
func Seven() error { // want `function "Seven" has odd docstring: an error code can't be purely whitespace`
	return &Error{"hello-error"}
}

// InvalidCodeFormat1 declares an error with invalid format.
//
// Errors:
//
// - invalid- -- ending with a dash
func InvalidCodeFormat1() error { // want `function "InvalidCodeFormat1" has odd docstring: declared error code has invalid format: should match .*`
	return nil
}

// InvalidCodeFormat2 declares an error with invalid format.
//
// Errors:
//
// - -invalid -- starting with a dash
func InvalidCodeFormat2() error { // want `function "InvalidCodeFormat2" has odd docstring: declared error code has invalid format: should match .*`
	return nil
}

// InvalidCodeFormat3 declares an error with invalid format.
//
// Errors:
//
// - 0invalid-error -- starting with a number
func InvalidCodeFormat3() error { // want `function "InvalidCodeFormat3" has odd docstring: declared error code has invalid format: should match .*`
	return nil
}

// InvalidCodeFormat4 declares an error with invalid format.
//
// Errors:
//
// - invalid(error)-code -- containing invalid chars (braces)
func InvalidCodeFormat4() error { // want `function "InvalidCodeFormat4" has odd docstring: declared error code has invalid format: should match .*`
	return nil
}

// InvalidCodeFormat5 declares an error with invalid format.
//
// Errors:
//
// - invalid error -- containing invalid char (space)
func InvalidCodeFormat5() error { // want `function "InvalidCodeFormat5" has odd docstring: declared error code has invalid format: should match .*`
	return nil
}

type InterfaceOne interface {
	One() error // want `interface method "One" does not declare any error codes`

	// OneWithComment is demo function.
	// Erros are not documented in this function which should be detected by our analyzer.
	OneWithComment() error // want `interface method "OneWithComment" does not declare any error codes`

	one() error // want `interface method "one" does not declare any error codes`
}

type OddDocstringInterface interface {
	// Two is a test function.
	//
	// Errors:
	//    - hello-error -- is always returned.
	Two() error // want `interface method "Two" has odd docstring: need a blank line after the 'Errors:' block indicator`

	// Three is a demo function.
	// The following errors docstring has multiple 'Errors:' block indicators which is invalid.
	//
	// Errors:
	//
	// Errors:
	//
	//    - hello-error -- is always returned.
	Three() error // want `interface method "Three" has odd docstring: repeated 'Errors:' block indicator`

	// Four is a demo function.
	// The following errors docstring has multiple 'Errors:' block indicators which is invalid.
	//
	// Errors:
	//
	//    - hello-error -- is always returned.
	//
	// Errors:
	//
	//    - hello-error -- is always returned.
	Four() error // want `interface method "Four" has odd docstring: repeated 'Errors:' block indicator`

	// Five is a demo function.
	// The following errors docstring has an error code line with an invalid format.
	//
	// Errors:
	//
	//    - hello-error - is always returned.
	Five() error // want `interface method "Five" has odd docstring: mid block, a line leading with '- ' didnt contain a '--' to mark the end of the code name`

	// Six is a demo function.
	// The following errors docstring has an invalid whitespace error code.
	//
	// Errors:
	//
	//    - hello-error -- is always returned.
	//    - -- is invalid.
	Six() error // want `interface method "Six" has odd docstring: an error code can't be purely whitespace`

	// Seven is a demo function.
	// The following errors docstring has an invalid whitespace error code.
	//
	// Errors:
	//
	//    - hello-error -- is always returned.
	//    -             -- is invalid.
	Seven() error // want `interface method "Seven" has odd docstring: an error code can't be purely whitespace`

	// InvalidCodeFormat1 declares an error with invalid format.
	//
	// Errors:
	//
	// - invalid- -- ending with a dash
	InvalidCodeFormat1() error // want `interface method "InvalidCodeFormat1" has odd docstring: declared error code has invalid format: should match .*`

	// InvalidCodeFormat2 declares an error with invalid format.
	//
	// Errors:
	//
	// - -invalid -- starting with a dash
	InvalidCodeFormat2() error // want `interface method "InvalidCodeFormat2" has odd docstring: declared error code has invalid format: should match .*`
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
