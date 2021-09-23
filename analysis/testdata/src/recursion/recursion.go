package recursion

// SimpleRecursion always returns an error,
// and looks to the compiler like it might do recursive calls on itself.
//
// Errors:
//
//    - recursion-error -- is always returned
func SimpleRecursion() error {
	if false {
		return SimpleRecursion()
	}
	return &Error{"recursion-error"}
}

// InnerRecursion always returns an error,
// and calls a local function, that looks recursive to the compiler.
//
// Errors:
//
//    - recursion-error -- is always returned
func InnerRecursion() error {
	return innerRecursion()
}

func innerRecursion() error {
	if false {
		return innerRecursion()
	}
	return &Error{"recursion-error"}
}

// IndirectRecursion always returns an error,
// and looks to the compiler like it might do an indirect recursion involving three functions.
//
// Errors:
//
//    - recursion-error       -- is always returned
//    - recursion-part1-error --
//    - recursion-part2-error --
func IndirectRecursion() error {
	if false {
		return part1()
	}
	return &Error{"recursion-error"}
}

func part1() error {
	if false {
		return part2()
	}
	return &Error{"recursion-part1-error"}
}

func part2() error {
	if false {
		return IndirectRecursion()
	}
	return &Error{"recursion-part2-error"}
}

// IndirectInnerRecursion always returns an error,
// and calls a local function, that looks to the compiler
// like it does indirect recursive calls involving three functions.
//
// Errors:
//
//    - recursion-inner1-error -- is always returned
//    - recursion-inner2-error --
//    - recursion-inner3-error --
func IndirectInnerRecursion() error {
	return inner1()
}

func inner1() error {
	if false {
		return inner2()
	}
	return &Error{"recursion-inner1-error"}
}

func inner2() error {
	if false {
		return inner3()
	}
	return &Error{"recursion-inner2-error"}
}

func inner3() error {
	if false {
		return inner1()
	}
	return &Error{"recursion-inner3-error"}
}

// IdentLoop does assignments in an attempt to make the analysis fall into an endless loop.
//
// Errors:
//
//    - some-error -- is always returned
func IdentLoop() error { // want Shadowed:"ErrorCodes: named-error"
	err1 := &Error{"some-error"}
	err2 := err1
	err3 := err2
	err1 = err3
	return err1
}

type Error struct {
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
