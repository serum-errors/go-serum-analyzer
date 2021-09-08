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
func Correct() error {
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

type Error struct {
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
