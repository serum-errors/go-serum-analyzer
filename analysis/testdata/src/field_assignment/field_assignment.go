package fieldassignment

// Errors:
//
//    - some-error  --
//    - other-error --
func EasyAssignment() error { // want EasyAssignment:"ErrorCodes: other-error some-error"
	err := &Error{"some-error"}
	if true {
		err.TheCode = "other-error"
	}
	return err
}

const errorSuffix = "-error"

// Errors:
//
//    - some-error    --
//    - other-1-error --
//    - other-2-error --
//    - other-3-error --
func ConstantExpressionAssignments() error { // want ConstantExpressionAssignments:"ErrorCodes: other-1-error other-2-error other-3-error some-error"
	err := Error{"some-error"}
	switch {
	case true:
		err.TheCode = "other-1" + errorSuffix
	case true:
		const code = "other-2-error"
		err.TheCode = code
	case true:
		const code1 = "other"
		const code2 = "-3"
		err.TheCode = code1 + code2 + errorSuffix
	}
	return &err
}

// Errors:
//
//    - some-error --
func AssignInvalidCode() error { // want AssignInvalidCode:"ErrorCodes: some-error"
	err := &Error{"some-error"}
	if true {
		err.TheCode = "-invalid-" // want "error code has invalid format: should match .*"
	}
	return err
}

func returnCode() string {
	return "func-error"
}

func returnTwoCodes() (string, string) {
	return "func-error", "other-func-error"
}

// Errors:
//
//    - some-error --
func AssignInvalidExpression(input string) error { // want AssignInvalidExpression:"ErrorCodes: some-error"
	err := &Error{"some-error"}
	switch {
	case true:
		err.TheCode = input // want "error code field has to be assigned a constant value"
	case true:
		err.TheCode = returnCode() // want "error code field has to be assigned a constant value"
	case true:
		_, err.TheCode = returnTwoCodes() // want "error code field has to be assigned a constant value"
	case true:
		err.TheCode, err.TheCode = returnTwoCodes() // want "error code field has to be assigned a constant value" "error code field has to be assigned a constant value"
	}
	return err
}

type OtherStuff struct {
	Field1, Field2 string
}

// Errors:
//
//    - some-error --
func AssignmentsToOtherStuff() error { // want AssignmentsToOtherStuff:"ErrorCodes: some-error"
	err := &Error{"some-error"}
	not_returned := &Error{"unused-error"}
	not_returned.TheCode = "other-unused-error"
	other := OtherStuff{"a", "b"}
	not_returned.TheCode = other.Field1 + other.Field2
	other.Field2 = not_returned.TheCode + "some garbage"
	return err
}

// Errors:
//
//    - some-error  --
//    - other-error --
func AssignmentToOtherField(input string) error { // want AssignmentToOtherField:"ErrorCodes: other-error some-error"
	err := &Error2{"stuff", "some-error"}
	switch {
	case true:
		err.Other = "constant"
	case true:
		err.Other = input
	case true:
		err.TheCode = "other-error"
	}
	return err
}

// Errors:
//
//    - some-error  --
//    - other-1-error --
//    - other-2-error --
func ComplicatedAssignment() error { // want ComplicatedAssignment:"ErrorCodes: other-1-error other-2-error some-error"
	err1 := &Error2{"other", "some-error"}
	err2 := err1
	err2.TheCode, err2.TheCode = "other-1-error", "other-2-error"
	err2.Other = "don't care"
	var err3 error
	err3 = err2
	_, err := "unused", err3
	return err
}

type Error struct { // want Error:`ErrorType{Field:{Name:"TheCode", Position:0}, Codes:}`
	TheCode string
}

func (e *Error) Code() string               { return e.TheCode }
func (e *Error) Message() string            { return e.TheCode }
func (e *Error) Details() map[string]string { return nil }
func (e *Error) Cause() error               { return nil }
func (e *Error) Error() string              { return e.Message() }

type Error2 struct { // want Error2:`ErrorType{Field:{Name:"TheCode", Position:1}, Codes:}`
	Other, TheCode string
}

func (e *Error2) Code() string               { return e.TheCode }
func (e *Error2) Message() string            { return e.TheCode }
func (e *Error2) Details() map[string]string { return nil }
func (e *Error2) Cause() error               { return nil }
func (e *Error2) Error() string              { return e.Message() }
