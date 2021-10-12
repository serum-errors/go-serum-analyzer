package interfaces

type InvalidSimpleImpl struct{}

// InterfaceMethod is a method used to implement SimpleInterface,
// but the declared (and actual) error codes are not a subset of the ones declared in the interface.
//
// Errors:
//
//    - unknown-error -- is always returned making InvalidSimpleImpl an invalid implementation of the interface
func (InvalidSimpleImpl) SimpleInterfaceMethod() error { // want SimpleInterfaceMethod:"ErrorCodes: unknown-error"
	return &Error{"unknown-error"}
}

func InvalidReturn() SimpleInterface {
	return InvalidSimpleImpl{} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
}

func InvalidReturn2() (int, int, SimpleInterface) {
	return 5, 42, InvalidSimpleImpl{} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
}

func InvalidReturn3() (_, _ int, _, _ SimpleInterface) {
	return 5, 42, nil, InvalidSimpleImpl{} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
}

func InvalidAssignment() *SimpleInterface {
	var x SimpleInterface
	x = InvalidSimpleImpl{} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	return &x
}

func InvalidAssignment2() {
	var x, y int
	var z SimpleInterface
	x, y, z = 7, 3, InvalidSimpleImpl{} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_, _, _ = x, y, z
}

func InvalidAssignment3() (x int, y, z SimpleInterface) {
	x = 5
	y = nil
	z = InvalidSimpleImpl{} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	return
}

func ReturnTwoInvalidSimpleImpl() (InvalidSimpleImpl, InvalidSimpleImpl) {
	return InvalidSimpleImpl{}, InvalidSimpleImpl{}
}

func InvalidAssignment4() {
	var y SimpleInterface
	x, y := ReturnTwoInvalidSimpleImpl() // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_, _ = x, y

	var u, v SimpleInterface
	u, v = ReturnTwoInvalidSimpleImpl() /*
		want
			`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
			`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]` */
	_, _ = u, v
}

func TakingSimpleInterface(_, _ int, _ SimpleInterface) {}

func TakingSimpleInterfaceVariadic(_ int, _ ...SimpleInterface) {}

func TakingSimpleInterfaceSlice(_ int, _ []SimpleInterface) {}

func InvalidFunctionCall() {
	param := InvalidSimpleImpl{}
	TakingSimpleInterface(42, 42, param) // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	TakingSimpleInterface(42, 42, nil)
	TakingSimpleInterfaceVariadic(-5, nil, param)      // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	TakingSimpleInterfaceVariadic(-5, param, nil, nil) // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	TakingSimpleInterfaceVariadic(-5, nil, nil, nil)

	// Make sure slice parameters don't get confused with variadic.
	TakingSimpleInterfaceSlice(7, []SimpleInterface{nil, nil})
	TakingSimpleInterfaceSlice(7, nil)

	var slice []SimpleInterface
	slice = append(slice, param)           // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	slice = append(slice, nil, nil, param) // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	slice = append(slice, nil, nil)

	lambda := func(_ SimpleInterface) {}
	lambda(InvalidSimpleImpl{}) // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
}

func InvalidMapIndex(m map[SimpleInterface]struct{}, slice []struct{}) {
	_ = m[InvalidSimpleImpl{}] // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = m[nil]
	_ = slice[7] // making sure indexing something other than a map does not fail
}

type SimpleInterfaceIncompatible interface { // want SimpleInterfaceIncompatible:"ErrorInterface: SimpleInterfaceMethod"
	// SimpleInterfaceMethod is a method returning an error with error codes declared in doc.
	//
	// Errors:
	//
	//    - interface-1-error    -- could potentially be returned
	//    - incompatible-1-error -- could potentially be returned
	//    - incompatible-2-error -- could potentially be returned
	SimpleInterfaceMethod() error // want SimpleInterfaceMethod:"ErrorCodes: incompatible-1-error incompatible-2-error interface-1-error"
}

func InvalidConversion(incompatible SimpleInterfaceIncompatible) {
	_ = SimpleInterface(InvalidSimpleImpl{}) // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = SimpleInterface(nil)
	_ = SimpleInterface(incompatible)       // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[incompatible-1-error incompatible-2-error]`
	_ = incompatible.(SimpleInterface)      // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[incompatible-1-error incompatible-2-error]`
	s, ok := incompatible.(SimpleInterface) // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[incompatible-1-error incompatible-2-error]`
	_, _ = s, ok
}

type BoxSimpleInterface struct {
	si SimpleInterface
}

type Box2SimpleInterface struct {
	_, _ int
	_    SimpleInterface
	_, _ SimpleInterface
}

type Box3SimpleInterface struct {
	u, v int
	x, y SimpleInterface
}

func InvalidStructCreation() {
	_ = BoxSimpleInterface{InvalidSimpleImpl{}} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = BoxSimpleInterface{nil}
	_ = BoxSimpleInterface{si: InvalidSimpleImpl{}} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = BoxSimpleInterface{si: nil}
	_ = BoxSimpleInterface{}
	_ = &BoxSimpleInterface{InvalidSimpleImpl{}} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`

	_ = Box2SimpleInterface{42, 5, InvalidSimpleImpl{}, nil, nil}                                 // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = Box2SimpleInterface{42, 5, nil, nil, InvalidSimpleImpl{}}                                 // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = Box2SimpleInterface{42, 5, InvalidSimpleImpl{}, InvalidSimpleImpl{}, InvalidSimpleImpl{}} /*
		want
			`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
			`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
			`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]` */
	_ = Box2SimpleInterface{}

	_ = Box3SimpleInterface{42, 5, nil, nil}
	_ = Box3SimpleInterface{42, 5, nil, InvalidSimpleImpl{}}             // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = Box3SimpleInterface{u: 42, v: 5, x: nil, y: InvalidSimpleImpl{}} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = Box3SimpleInterface{u: 42, x: nil, y: InvalidSimpleImpl{}}       // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = Box3SimpleInterface{u: 42}

	_ = []BoxSimpleInterface{
		{si: InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		{InvalidSimpleImpl{}},     // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		{si: nil},
		{nil},
		{},
	}

	_ = struct{ si SimpleInterface }{si: InvalidSimpleImpl{}} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = struct{ si SimpleInterface }{InvalidSimpleImpl{}}     // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = struct{ si SimpleInterface }{si: nil}
	_ = struct{ si SimpleInterface }{nil}
	_ = struct{ si SimpleInterface }{}

	_ = []struct{ si SimpleInterface }{
		{si: InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		{InvalidSimpleImpl{}},     // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		{si: nil},
		{nil},
		{},
	}

	_ = map[string]BoxSimpleInterface{
		"a": {si: InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		"b": {InvalidSimpleImpl{}},     // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		"c": {si: nil},
		"d": {nil},
		"e": {},
	}
}

type interfaceSlice []SimpleInterface
type interfaceSlice2 interfaceSlice

func InvalidSliceCreation() {
	_ = []SimpleInterface{}
	_ = []SimpleInterface{nil, nil, nil}
	_ = []SimpleInterface{nil, InvalidSimpleImpl{}, nil} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = []SimpleInterface{
		nil,
		InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		nil,
		nil,
		InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	}

	const index = 40 + 2
	_ = []SimpleInterface{
		4:     nil,
		1:     InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		2:     nil,
		3:     nil,
		300:   InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		index: InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	}

	_ = interfaceSlice{}
	_ = interfaceSlice{nil}
	_ = interfaceSlice{InvalidSimpleImpl{}, nil}       // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = interfaceSlice{1: InvalidSimpleImpl{}, 0: nil} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`

	_ = interfaceSlice2{}
	_ = interfaceSlice2{nil}
	_ = interfaceSlice2{InvalidSimpleImpl{}, nil}       // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = interfaceSlice2{1: InvalidSimpleImpl{}, 0: nil} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`

	_ = [][][]SimpleInterface{
		{
			{nil, InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
			{},
			{nil},
		},
		{
			7:                     {},
			8:                     {8: InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
			{InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		},
	}
	_ = []interfaceSlice2{{nil, nil, InvalidSimpleImpl{}}} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
}

type interfaceArray [2]SimpleInterface
type interfaceArray2 interfaceSlice

func InvalidArrayCreation() {
	_ = [0]SimpleInterface{}
	_ = [3]SimpleInterface{nil, nil, nil}
	_ = [3]SimpleInterface{nil, InvalidSimpleImpl{}, nil} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = [...]SimpleInterface{
		nil,
		InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		nil,
		nil,
		InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	}

	const index = 40 + 2
	_ = [...]SimpleInterface{
		4:     nil,
		1:     InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		2:     nil,
		3:     nil,
		300:   InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		index: InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	}

	_ = interfaceArray{}
	_ = interfaceArray{nil}
	_ = interfaceArray{InvalidSimpleImpl{}, nil}       // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = interfaceArray{1: InvalidSimpleImpl{}, 0: nil} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`

	_ = interfaceArray2{}
	_ = interfaceArray2{nil}
	_ = interfaceArray2{InvalidSimpleImpl{}, nil}       // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = interfaceArray2{1: InvalidSimpleImpl{}, 0: nil} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`

	_ = [...][10][10]SimpleInterface{
		{
			{nil, InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
			{},
			{nil},
		},
		{
			7:                     {},
			8:                     {8: InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
			{InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		},
	}
	_ = []interfaceArray2{{nil, InvalidSimpleImpl{}}} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
}

type interfaceMap map[int]SimpleInterface
type interfaceMap2 interfaceSlice

func InvalidMapCreation() {
	_ = map[string]SimpleInterface{}
	_ = map[string]SimpleInterface{"a": nil, "b": nil}
	_ = map[string]SimpleInterface{"x": nil, "y": InvalidSimpleImpl{}, "z": nil} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`

	const index = 40 + 2
	_ = map[int]SimpleInterface{
		4:     nil,
		1:     InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		2:     nil,
		3:     nil,
		300:   InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		index: InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	}

	_ = interfaceMap{}
	_ = interfaceMap{3: nil}
	_ = interfaceMap{1: InvalidSimpleImpl{}, 0: nil} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`

	_ = interfaceMap2{}
	_ = interfaceMap2{7: nil}
	_ = interfaceMap2{1: InvalidSimpleImpl{}, 0: nil} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`

	_ = map[string]map[int]map[int]SimpleInterface{
		"a": {
			0: {4: nil, 42: InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
			1: {},
			2: {5: nil},
		},
		"b": {
			7: {},
			8: {8: InvalidSimpleImpl{}}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		},
	}
	_ = map[string]interfaceMap2{"index": {0: nil, 5: InvalidSimpleImpl{}}} // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`

	_ = map[SimpleInterface]int{}
	_ = map[SimpleInterface]int{
		InvalidSimpleImpl{}: 5, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	}
	_ = map[SimpleInterface]int{
		nil:                 42,
		InvalidSimpleImpl{}: 5, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		nil:                 3,
	}

	_ = map[SimpleInterface]SimpleInterface{
		nil:                 InvalidSimpleImpl{}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		InvalidSimpleImpl{}: nil,                 // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		InvalidSimpleImpl{}: InvalidSimpleImpl{}, /*
			want
				`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
				`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]` */
	}

	_ = map[string][]SimpleInterface{
		"index": {InvalidSimpleImpl{}, nil}, // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	}
}
