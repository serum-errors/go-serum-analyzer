package examples

type Collection struct {
	values []interface{}
	limit  int
}

// Add adds a non-nil value into the collection.
//
// Errors:
//
//    - examples-error-invalid-arg        -- if the given argument is nil
//    - examples-error-invalid-collection -- if the given collection is nil or invalid
//    - examples-error-limit-reached      -- if the limit of values in the collection is reached
func (c *Collection) Add(item interface{}) error { // want Add:"ErrorCodes: examples-error-invalid-arg examples-error-invalid-collection examples-error-limit-reached"
	if item == nil {
		return &Error{"examples-error-invalid-arg"}
	}

	if c == nil || c.limit < 0 {
		return &Error{"examples-error-invalid-collection"}
	}

	if len(c.values) >= c.limit {
		return &Error{"examples-error-limit-reached"}
	}

	c.values = append(c.values, item)
	return nil
}

// Errors:
//
// None of the following errors are actually returned yet.
//    - examples-error-invalid-arg        -- if the given argument is nil
//    - examples-error-invalid-collection -- if the given collection is nil or invalid
//    - examples-error-limit-reached      -- if the limit of values in the collection is reached
func (c *Collection) AddUnused(item interface{}) error { /*
		want
			AddUnused:"ErrorCodes: examples-error-invalid-arg examples-error-invalid-collection examples-error-limit-reached"
			`function "AddUnused" has a mismatch of declared and actual error codes: unused codes: \[examples-error-invalid-arg examples-error-invalid-collection examples-error-limit-reached]` */
	panic("not implemented")
}

// Errors: none -- not actually true, but we want to showcasse missing error codes.
func (c *Collection) AddMissing(item interface{}) error { /*
		want
			AddMissing:"ErrorCodes:"
			`function "AddMissing" has a mismatch of declared and actual error codes: missing codes: \[examples-error-invalid-arg examples-error-invalid-collection examples-error-limit-reached]` */
	if item == nil {
		return &Error{"examples-error-invalid-arg"}
	}

	if c == nil || c.limit < 0 {
		return &Error{"examples-error-invalid-collection"}
	}

	if len(c.values) >= c.limit {
		return &Error{"examples-error-limit-reached"}
	}

	c.values = append(c.values, item)
	return nil
}

// AddAlt adds a non-nil value into the collection.
//
// Errors:
//
//    - examples-error-invalid-arg        -- if the given argument is nil
//    - examples-error-invalid-collection -- if the given collection is nil or invalid
//    - examples-error-limit-reached      -- if the limit of values in the collection is reached
func (c *Collection) AddAlt(item interface{}) error { // want AddAlt:"ErrorCodes: examples-error-invalid-arg examples-error-invalid-collection examples-error-limit-reached"
	var err error

	switch {
	case item == nil:
		err = &Error{"examples-error-invalid-arg"}
	case c == nil || c.limit < 0:
		err = &Error{"examples-error-invalid-collection"}
	case len(c.values) >= c.limit:
		err = &Error{"examples-error-limit-reached"}
	default:
		c.values = append(c.values, item)
	}

	return err
}

type (
	IO interface { // want IO:"ErrorInterface: Read"
		// Errors:
		//
		//    - examples-error-io         -- if a general io error occurs
		//    - examples-error-disconnect -- if the peer disconnected
		Read() (byte, error) // want Read:"ErrorCodes: examples-error-disconnect examples-error-io"
	}
	MockedIO struct{}
)

// Read for MockedIO always returns 0.
// Errors: none -- this method only returns error to comply with the interface IO.
func (MockedIO) Read() (byte, error) { return 0, nil } // want Read:"ErrorCodes:"
