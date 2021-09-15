package errortypes

// AllErrors returns variations of all errors defined in this package.
//
// Errors:
//
//    - some-error       --
//    - some-2-error     --
//    - some-3-error     --
//    - some-4-error     --
//    - multiple-1-error --
//    - multiple-2-error --
//    - multiple-3-error --
//    - value-1-error    --
//    - value-2-error    --
//    - field-1-error    --
//    - field-2-error    --
//    - field-3-error    --
//    - field-4-error    --
//    - field-5-error    --
//    - field-6-error    --
//    - promoted-1-error --
//    - promoted-2-error --
//    - promoted-3-error --
//    - combined-1-error --
//    - combined-2-error --
//    - combined-3-error --
func AllErrors(param1 string) error { // want AllErrors:"ErrorCodes: combined-1-error combined-2-error combined-3-error field-1-error field-2-error field-3-error field-4-error field-5-error field-6-error multiple-1-error multiple-2-error multiple-3-error promoted-1-error promoted-2-error promoted-3-error some-2-error some-3-error some-4-error some-error value-1-error value-2-error"
	switch {
	case true:
		return &ConstantError{}
	case true:
		return &ConstantError2{}
	case true:
		return &ConstantError3{}
	case true:
		return &ConstantError4{}
	case true:
		return &MultipleConstantError{}
	case true:
		return ValueTypeError{}
	case true:
		return &ValueTypeError2{} // valid because methods of T are also methods of *T
	case true:
		return &InvalidError{}
	case true:
		return &InvalidError2{}
	case true:
		return &FieldError{"field-1-error"}
	case true:
		return &FieldError2{"field-2-error", "some other", "values"}
	case true:
		return &FieldError3{"something", "field-3-error", "something else"}
	case true:
		return &FieldError{"field-4-error"} // repeated FieldError to test if multiple error codes can originate using the same type
	case true:
		return &FieldError{field: "field-5-error"} // simple test for named constructor
	case true:
		return &FieldError2{field3: "unrelated", field2: "stuff", field1: "field-6-error"} // more advanced test for named constructor
	case true:
		return &FieldError{param1} // want "error code field has to be instantiated by constant value"
	case true:
		return &FieldError{} // want "error code field has to be instantiated by constant value"
	case true:
		return &FieldError{"badformat-"} // want "error code has invalid format: should match .*"
	case true:
		return &PromotedFieldError{Promoteable{"one", "two"}, "three", "promoted-1-error"}
	case true:
		return &PromotedFieldError2{field: "promoted-2-error"}
	case true:
		return &PromotedFieldError3{nil, "promoted-3-error", "something"}
	case true:
		return &InvalidPromotedFieldError{Promoteable{"x", "y"}}
	case true:
		return &CombinedError{"combined-3-error"}
	}
	return nil
}

type ConstantError struct{} // want ConstantError:"ErrorType{Field:<nil>, Codes:some-error}"

func (e *ConstantError) Code() string  { return "some-error" }
func (e *ConstantError) Error() string { return "ConstantError" }

type ConstantError2 struct{} // want ConstantError2:"ErrorType{Field:<nil>, Codes:some-2-error}"

func (e *ConstantError2) Code() string  { return "some-2" + "-error" }
func (e *ConstantError2) Error() string { return "ConstantError2" }

const constantError3Code = "some-3" + "-error"

type ConstantError3 struct{} // want ConstantError3:"ErrorType{Field:<nil>, Codes:some-3-error}"

func (e *ConstantError3) Code() string  { return constantError3Code }
func (e *ConstantError3) Error() string { return "ConstantError3" }

type ConstantError4 struct{} // want ConstantError4:"ErrorType{Field:<nil>, Codes:some-4-error}"

func (*ConstantError4) Code() string  { return "some-4-error" }
func (*ConstantError4) Error() string { return "ConstantError4" }

type MultipleConstantError struct{} // want MultipleConstantError:"ErrorType{Field:<nil>, Codes:multiple-1-error multiple-2-error multiple-3-error}"

func (e *MultipleConstantError) Code() string {
	switch {
	case true:
		return "multiple-1-error"
	case true:
		return "multiple-2-error"
	default:
		return "multiple-3-error"
	}
}
func (e *MultipleConstantError) Error() string { return "MultipleConstantError" }

type ValueTypeError struct{} // want ValueTypeError:"ErrorType{Field:<nil>, Codes:value-1-error}"

func (e ValueTypeError) Code() string  { return "value-1-error" }
func (e ValueTypeError) Error() string { return "ValueTypeError" }

type ValueTypeError2 struct{} // want ValueTypeError2:"ErrorType{Field:<nil>, Codes:value-2-error}"

func (e ValueTypeError2) Code() string  { return "value-2-error" }
func (e ValueTypeError2) Error() string { return "ValueTypeError2" }

type InvalidError struct{}

func (e *InvalidError) Code() string  { return "invalid error" } // want "error code has invalid format: should match .*"
func (e *InvalidError) Error() string { return "InvalidError" }

type InvalidError2 struct{ field1, field2 string }

func (e *InvalidError2) Code() string  { return e.field1 + e.field2 } // want `function "Code" should always return a string constant or a single field`
func (e *InvalidError2) Error() string { return "InvalidError2" }

type FieldError struct{ field string } // want FieldError:`ErrorType{Field:{Name:"field", Position:0}, Codes:}`

func (e *FieldError) Code() string  { return e.field }
func (e *FieldError) Error() string { return "FieldError" }

type FieldError2 struct{ field1, field2, field3 string } // want FieldError2:`ErrorType{Field:{Name:"field1", Position:0}, Codes:}`

func (e *FieldError2) Code() string {
	switch {
	case true:
		return e.field1
	case true:
		return e.field2 // want `only single field allowed: cannot return field "field2" because field .* was returned previously`
	default:
		return e.field3 // want `only single field allowed: cannot return field "field3" because field .* was returned previously`
	}
}
func (e *FieldError2) Error() string { return "FieldError2" }

type FieldError3 struct{ _, field2, _ string } // want FieldError3:`ErrorType{Field:{Name:"field2", Position:1}, Codes:}`

func (e *FieldError3) Code() string  { return e.field2 }
func (e *FieldError3) Error() string { return "FieldError3" }

type Promoteable struct{ Some, Other string }
type Promoteable2 struct{ _, _, _ int }
type Promoteable3 struct{}

type PromotedFieldError struct { // want PromotedFieldError:`ErrorType{Field:{Name:"field", Position:2}, Codes:}`
	Promoteable
	_, field string
}

func (e *PromotedFieldError) Code() string  { return e.field }
func (e *PromotedFieldError) Error() string { return "PromotedFieldError" }

type PromotedFieldError2 struct { // want PromotedFieldError2:`ErrorType{Field:{Name:"field", Position:4}, Codes:}`
	Promoteable
	Promoteable2
	_        int
	_, field string
	Promoteable3
}

func (e *PromotedFieldError2) Code() string  { return e.field }
func (e *PromotedFieldError2) Error() string { return "PromotedFieldError2" }

type PromotedFieldError3 struct { // want PromotedFieldError3:`ErrorType{Field:{Name:"errorCode", Position:1}, Codes:}`
	*Promoteable
	errorCode, _ string
}

func (e *PromotedFieldError3) Code() string  { return e.errorCode }
func (e *PromotedFieldError3) Error() string { return "PromotedFieldError3" }

type InvalidPromotedFieldError struct{ Promoteable }

func (e *InvalidPromotedFieldError) Code() string  { return e.Some } // want `returned field "Some" is not a valid error code field`
func (e *InvalidPromotedFieldError) Error() string { return "InvalidPromotedFieldError" }

type CombinedError struct{ field string } // want CombinedError:`ErrorType{Field:{Name:"field", Position:0}, Codes:combined-1-error combined-2-error}`

func (e *CombinedError) Code() string {
	switch {
	case true:
		return e.field
	case true:
		return "combined-1-error"
	}
	return "combined-2-error"
}
func (e *CombinedError) Error() string { return "CombinedError" }
