package interfaces

type InvalidSimpleImpl struct{}

type InterFaceWithoutErrorCodeDeclarations interface {
	ExportedCodeNotDeclared() error // want `interface method "ExportedCodeNotDeclared" does not declare any error codes`
	localCodeNotDeclared() error    // want `interface method "localCodeNotDeclared" does not declare any error codes`
}

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

func InvalidAssignment5() {
	var x1 SimpleInterface = InvalidSimpleImpl{}                  // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	var x2, x3, _ SimpleInterface = nil, InvalidSimpleImpl{}, nil // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	var x4, _ SimpleInterface = ReturnTwoInvalidSimpleImpl()      /*
		want
			`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
			`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]` */
	_, _, _, _ = x1, x2, x3, x4
}

func InvalidAssignment6(m map[int]InvalidSimpleImpl) {
	var si SimpleInterface
	var ok bool
	if si = m[0]; true { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}
	if si, ok = m[0]; ok { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}

	for si, ok = m[0]; ok; { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}

	switch si, ok = m[0]; { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	case ok:
		_ = si
	}
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

type RandomInterface interface {
	RandomMethod()
}

func InvalidConversion(incompatible SimpleInterfaceIncompatible, bs RandomInterface) {
	_ = SimpleInterface(InvalidSimpleImpl{}) // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	_ = SimpleInterface(nil)
	_ = SimpleInterface(incompatible)       // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[incompatible-1-error incompatible-2-error]`
	_ = incompatible.(SimpleInterface)      // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[incompatible-1-error incompatible-2-error]`
	s, ok := incompatible.(SimpleInterface) // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[incompatible-1-error incompatible-2-error]`
	_, _ = s, ok

	var si SimpleInterface
	switch val := incompatible.(type) {
	case RandomInterface:
		_ = val
	case SimpleInterface: // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[incompatible-1-error incompatible-2-error]`
		si = val
	default:
		_ = val
	}
	_ = si
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

func InvalidForRange(
	slice []InvalidSimpleImpl,
	m1 map[InvalidSimpleImpl]struct{},
	m2 map[string]InvalidSimpleImpl,
	m3 map[InvalidSimpleImpl]InvalidSimpleImpl,
	array [5]InvalidSimpleImpl,
	pa *[5]InvalidSimpleImpl,
	c chan InvalidSimpleImpl,
) {
	var si SimpleInterface
	for _, si = range slice { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}

	for i := range slice {
		_ = i
	}

	for _, si = range array { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}

	for i := range array {
		_ = i
	}

	for _, si = range pa { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}

	for i := range pa {
		_ = i
	}

	for si, _ = range m1 { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}

	for si = range m1 { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}

	for _, si = range m2 { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}

	var si2 SimpleInterface
	for si, si2 = range m3 { /*
			want
				`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
				`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]` */
		_, _ = si, si2
	}

	for si, si = range m3 { /*
			want
				`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
				`cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]` */
		_ = si
	}

	for si = range m3 { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}

	for si = range c { // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	}
}

func InvalidChannelOperation(c chan InvalidSimpleImpl, csi chan SimpleInterface) {
	var si SimpleInterface = <-c // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	var some = <-c               // no problem
	c <- InvalidSimpleImpl{}     // no problem
	c <- some                    // no problem
	csi <- InvalidSimpleImpl{}   // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	csi <- <-c                   // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
	csi <- si                    // no problem

	// Same cases but this time in a select statement:
	select {
	case si = <-c: // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	case some = <-c:
		_ = si
	case c <- InvalidSimpleImpl{}:
		_ = si
	case c <- some:
		_ = si
	case csi <- InvalidSimpleImpl{}: // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	case csi <- <-c: // want `cannot use expression as "SimpleInterface" value: method "SimpleInterfaceMethod" declares the following error codes which were not part of the interface: \[unknown-error]`
		_ = si
	case csi <- si:
		_ = si
	}
}

type (
	someLocalInterface interface { // want someLocalInterface:"ErrorInterface: local1 local2"
		// Errors:
		//
		//    - local-1-error --
		//    - local-2-error --
		local1() error // want local1:"ErrorCodes: local-1-error local-2-error"

		// Errors:
		//
		//    - local-3-error --
		//    - local-4-error --
		local2() error // want local2:"ErrorCodes: local-3-error local-4-error"

		local3() error // want `interface method "local3" does not declare any error codes`
	}

	someLocalInterfaceInvalidImpl struct{}
)

func (someLocalInterfaceInvalidImpl) local1() error {
	return &Error{"unknown-error"}
}

func (someLocalInterfaceInvalidImpl) local2() error {
	switch {
	case true:
		return &Error{"some-1-error"}
	case true:
		return &Error{"local-3-error"} // this one is ok
	}
	return &Error{"some-2-error"}
}

func (someLocalInterfaceInvalidImpl) local3() error {
	return &Error{"unknown-error"}
}

func InvalidLocalImplementation() {
	var local someLocalInterface = someLocalInterfaceInvalidImpl{} /*
		want
			`cannot use expression as "someLocalInterface" value: method "local1" declares the following error codes which were not part of the interface: \[unknown-error]`
			`cannot use expression as "someLocalInterface" value: method "local2" declares the following error codes which were not part of the interface: \[some-1-error some-2-error]` */
	_ = local
}

type (
	Interface1 interface { // want Interface1:"ErrorInterface: I1"
		// Errors: none
		I1() error // want I1:"ErrorCodes:"
	}
	Interface2 interface { // want Interface2:"ErrorInterface: I2"
		// Errors: none
		I2() error // want I2:"ErrorCodes:"
	}
	Interface3 interface { // want Interface3:"ErrorInterface: I3"
		// Errors: none
		I3() error // want I3:"ErrorCodes:"
	}
	Interface1InvalidImpl struct {
		_ string
		_ Interface2
	}
	Interface2InvalidImpl struct{}
	Interface3InvalidImpl struct{}
)

// Errors:
//
//    - unknown-error --
func (Interface1InvalidImpl) I1() error { // want I1:"ErrorCodes: unknown-error"
	return &Error{"unknown-error"}
}

// Errors:
//
//    - unknown-error --
func (Interface2InvalidImpl) I2() error { // want I2:"ErrorCodes: unknown-error"
	return &Error{"unknown-error"}
}

// Errors:
//
//    - unknown-1-error --
//    - unknown-2-error --
//    - unknown-3-error --
func (Interface3InvalidImpl) I3() error { // want I3:"ErrorCodes: unknown-1-error unknown-2-error unknown-3-error"
	switch {
	case true:
		return &Error{"unknown-1-error"}
	case true:
		return &Error{"unknown-2-error"}
	default:
		return &Error{"unknown-3-error"}
	}
}

func GetInterface2InvalidImpl(_, _ int, _ Interface3) Interface2InvalidImpl {
	return Interface2InvalidImpl{}
}

func NestedInvalidConversions(i3 Interface3InvalidImpl) {
	var i1 Interface1 = Interface1InvalidImpl{"some string", GetInterface2InvalidImpl(1, 2, i3)} /*
		want
			`cannot use expression as "Interface1" value: method "I1" declares the following error codes which were not part of the interface: \[unknown-error]`
			`cannot use expression as "Interface2" value: method "I2" declares the following error codes which were not part of the interface: \[unknown-error]`
			`cannot use expression as "Interface3" value: method "I3" declares the following error codes which were not part of the interface: \[unknown-1-error unknown-2-error unknown-3-error]` */
	_ = i1
}
