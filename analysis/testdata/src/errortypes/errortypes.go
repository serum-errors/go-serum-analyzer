package errortypes

// AllErrors returns variations of all errors defined in this package.
//
// Errors:
//
//    - some-error       --
//    - some-2-error     --
//    - some-3-error     --
//    - multiple-1-error --
//    - multiple-2-error --
//    - multiple-3-error --
//    - value-1-error    --
//    - value-2-error    --
func AllErrors() error { // want AllErrors:"ErrorCodes: multiple-1-error multiple-2-error multiple-3-error some-2-error some-3-error some-error value-1-error value-2-error"
	switch {
	case true:
		return &ConstantError{}
	case true:
		return &ConstantError2{}
	case true:
		return &ConstantError3{}
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
	}
	return nil
}

type ConstantError struct{}

func (e *ConstantError) Code() string  { return "some-error" }
func (e *ConstantError) Error() string { return "ConstantError" }

type ConstantError2 struct{}

func (e *ConstantError2) Code() string  { return "some-2" + "-error" }
func (e *ConstantError2) Error() string { return "ConstantError2" }

const constantError3Code = "some-3" + "-error"

type ConstantError3 struct{}

func (e *ConstantError3) Code() string  { return constantError3Code }
func (e *ConstantError3) Error() string { return "ConstantError3" }

type MultipleConstantError struct{}

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

type ValueTypeError struct{}

func (e ValueTypeError) Code() string  { return "value-1-error" }
func (e ValueTypeError) Error() string { return "ValueTypeError" }

type ValueTypeError2 struct{}

func (e ValueTypeError2) Code() string  { return "value-2-error" }
func (e ValueTypeError2) Error() string { return "ValueTypeError2" }

type InvalidError struct{}

func (e *InvalidError) Code() string  { return "invalid error" } // want "error code from expression has invalid format: should match .*"
func (e *InvalidError) Error() string { return "InvalidError" }

type InvalidError2 struct{ field1, field2 string }

func (e *InvalidError2) Code() string  { return e.field1 + e.field2 } // want `function "Code" should always return a string constant or a single field`
func (e *InvalidError2) Error() string { return "InvalidError2" }
