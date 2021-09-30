package recursion

// SimpleRecursion always returns an error,
// and looks to the compiler like it might do recursive calls on itself.
//
// Errors:
//
//    - recursion-error -- is always returned
func SimpleRecursion() error { // want SimpleRecursion:"ErrorCodes: recursion-error"
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
func InnerRecursion() error { // want InnerRecursion:"ErrorCodes: recursion-error"
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
func IndirectRecursion() error { // want IndirectRecursion:"ErrorCodes: recursion-error recursion-part1-error recursion-part2-error"
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
func IndirectInnerRecursion() error { // want IndirectInnerRecursion:"ErrorCodes: recursion-inner1-error recursion-inner2-error recursion-inner3-error"
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

// Errors:
//
//    - advanced-a-error -- is always returned
//    - advanced-c-error --
//    - advanced-d-error --
//    - advanced-e-error --
//    - advanced-f-error --
//    - advanced-g-error --
func AdvancedRecursionA() error { // want AdvancedRecursionA:"ErrorCodes: advanced-a-error advanced-c-error advanced-d-error advanced-e-error advanced-f-error advanced-g-error"
	if false {
		return advancedRecursionC()
	}
	return &Error{"advanced-a-error"}
}

// Errors:
//
//    - advanced-b-error -- is always returned
//    - advanced-c-error --
//    - advanced-d-error --
//    - advanced-e-error --
//    - advanced-f-error --
//    - advanced-g-error --
func AdvancedRecursionB() error { // want AdvancedRecursionB:"ErrorCodes: advanced-b-error advanced-c-error advanced-d-error advanced-e-error advanced-f-error advanced-g-error"
	if false {
		return advancedRecursionD()
	}
	return &Error{"advanced-b-error"}
}

func advancedRecursionC() error {
	switch {
	case true:
		return advancedRecursionC()
	case true:
		return advancedRecursionD()
	case true:
		return advancedRecursionE()
	}
	return &Error{"advanced-c-error"}
}

func advancedRecursionD() error {
	if false {
		return advancedRecursionC()
	}
	return &Error{"advanced-d-error"}
}

func advancedRecursionE() error {
	if false {
		return advancedRecursionF()
	}
	return &Error{"advanced-e-error"}
}

func advancedRecursionF() error {
	if false {
		return advancedRecursionG()
	}
	return &Error{"advanced-f-error"}
}

func advancedRecursionG() error {
	if false {
		return advancedRecursionE()
	}
	return &Error{"advanced-g-error"}
}

// Errors:
//
//    - advanced-h-error -- is always returned
//    - advanced-e-error --
//    - advanced-f-error --
//    - advanced-g-error --
func AdvancedRecursionH() error { // want AdvancedRecursionH:"ErrorCodes: advanced-e-error advanced-f-error advanced-g-error advanced-h-error"
	if false {
		return advancedRecursionG()
	}
	return &Error{"advanced-h-error"}
}

// Errors:
//
//    - splitmerge-a-error --
//    - splitmerge-b-error --
//    - splitmerge-c-error --
//    - splitmerge-d-error --
//    - splitmerge-e-error --
func SplitMergeA() error { // want SplitMergeA:"ErrorCodes: splitmerge-a-error splitmerge-b-error splitmerge-c-error splitmerge-d-error splitmerge-e-error"
	switch {
	case true:
		return splitMergeB()
	case true:
		return splitMergeC()
	}
	return &Error{"splitmerge-a-error"}
}

func splitMergeB() error {
	if false {
		return splitMergeD()
	}
	return &Error{"splitmerge-b-error"}
}

func splitMergeC() error {
	if false {
		return splitMergeD()
	}
	return &Error{"splitmerge-c-error"}
}

func splitMergeD() error {
	if false {
		return splitMergeE()
	}
	return &Error{"splitmerge-d-error"}
}

func splitMergeE() error {
	if false {
		return splitMergeD()
	}
	return &Error{"splitmerge-e-error"}
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }
