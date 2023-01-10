package typeassertion

// Errors:
//
//  - interface-error --
func TypeAssertionMethod() *Error { // want TypeAssertionMethod:"ErrorCodes: interface-error"
	value := Receiver{}.Method()
	return value.(*Error)
}

// Errors:
//
//  - empty-interface-error --
func TypeAssertionInterface() error { // want TypeAssertionInterface:"ErrorCodes: empty-interface-error"
	return EmptyInterfaceError().(error)
}

// Errors:
//
// - standard-error --
func TypeAssertion() *Error { // want TypeAssertion:"ErrorCodes: standard-error"
	return StandardError().(*Error)
}

func TypeAssertionNoCodes() *Error { // want `function "TypeAssertionNoCodes" is exported, but does not declare any error codes`
	return StandardError().(*Error)
}

// Errors:
//
// - maybe-error -- sometimes
// - standard-error -- never but the analyser doesn't know that
func TypeAssertionMultipleTypes() error { // want TypeAssertionMultipleTypes:"ErrorCodes: maybe-error standard-error"
	if err := MaybeError(); err != nil {
		if err, ok := err.(ErrorInterface); ok {
			return err
		}
		return StandardError()
	}
	return nil
}

// Errors:
//
// - maybe-error -- sometimes
// - standard-error -- never but the analyser doesn't know that
func TypeAssertionSwitch() error { // want TypeAssertionSwitch:"ErrorCodes: maybe-error standard-error"
	err := MaybeError()
	switch x := err.(type) {
	case *Error:
		return x
	case nil:
		return nil
	default:
		return StandardError()
	}
}

// Errors:
//
//   - empty-interface-error -- always
//   - standard-error -- never
func TypeAssertionSwitchFromNonError() error { // want TypeAssertionSwitchFromNonError:"ErrorCodes: empty-interface-error standard-error"
	iface := EmptyInterfaceError()
	switch err := iface.(type) {
	case *Error:
		return err
	case ErrorInterface:
		return err
	case nil:
		return nil
	default:
		return StandardError()
	}
}
