package examples

type Box interface { // want Box:"ErrorInterface: Pop Put"
	// Put makes the box store the given value.
	//
	// Errors:
	//
	//    - examples-error-arg-nil -- if the given value is nil
	//    - examples-error-invalid -- if the box already holds a value
	//    - examples-error-unknown -- if an unexpected error occurred
	Put(value interface{}) error // want Put:"ErrorCodes: examples-error-arg-nil examples-error-invalid"

	// Pop retrieves the value stored in the box and removes it from the box.
	//
	// Errors:
	//
	//    - examples-error-invalid -- if the box was empty
	Pop() (interface{}, error) // want Pop:"ErrorCodes: examples-error-invalid"
}

type BoxImpl struct {
	value interface{}
}

// Errors:
//
//    - examples-error-arg-nil -- if the given value is nil
//    - examples-error-invalid -- if the box already holds a value
func (b *BoxImpl) Put(value interface{}) error { // want Put:"ErrorCodes: examples-error-arg-nil examples-error-invalid"
	if value == nil {
		return &Error{"examples-error-arg-nil"}
	}

	if b == nil || b.value != nil {
		return &Error{"examples-error-invalid"}
	}

	b.value = value
	return nil
}

// Errors:
//
//    - examples-error-invalid -- if the box was empty
func (b *BoxImpl) Pop() (interface{}, error) { // want Pop:"ErrorCodes: examples-error-invalid"
	if b == nil || b.value == nil {
		return nil, &Error{"examples-error-invalid"}
	}

	b.value = nil
	return b.value, nil
}

// UseBoxImplAsBox converts BoxImpl to the interface Box,
// showing that it can be used that way.
func UseBoxImplAsBox() {
	var b Box = &BoxImpl{}
	b.Put(b)
}

type BoxInvalidImpl struct{}

// Errors:
//
//    - examples-error-not-implemented --
func (b *BoxInvalidImpl) Put(value interface{}) error { // want Put:"ErrorCodes: examples-error-not-implemented"
	return &Error{"examples-error-not-implemented"}
}

// Errors:
//
//    - examples-error-not-implemented --
func (b *BoxInvalidImpl) Pop() (interface{}, error) { // want Pop:"ErrorCodes: examples-error-not-implemented"
	return nil, &Error{"examples-error-not-implemented"}
}

// UseBoxInvalidImplAsBox converts BoxInvalidImpl to the interface Box,
// showing that the analyser flags this as a problem.
func UseBoxInvalidImplAsBox() {
	var b Box = &BoxInvalidImpl{} /*
		want
			`cannot use expression as "Box" value: method "Put" declares the following error codes which were not part of the interface: \[examples-error-not-implemented]`
			`cannot use expression as "Box" value: method "Pop" declares the following error codes which were not part of the interface: \[examples-error-not-implemented]` */
	b.Put(b)
}

type Box2 interface { // want Box2:"ErrorInterface: Pop Put"
	// Put makes the box store the given value.
	//
	// Errors: none
	Put(value interface{}) error // want Put:"ErrorCodes:"

	// Pop retrieves the value stored in the box and removes it from the box.
	//
	// Errors:
	//
	//    - examples-error-invalid -- in case of an invalid operation
	Pop() (interface{}, error) // want Pop:"ErrorCodes: examples-error-invalid"
}

type EmbeddingBox interface { // want EmbeddingBox:"ErrorInterface: Pop Put"
	Box
	Box2 // want `embedded interface is not compatible: method "Put" has mismatches in declared error codes: missing codes: \[examples-error-arg-nil examples-error-invalid examples-error-unknown]`
}
