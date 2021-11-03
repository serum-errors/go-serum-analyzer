package interfaces

type (
	// Test if the check works in a multifile scenario
	EmbeddedMultipleInterfaces2 interface { // want EmbeddedMultipleInterfaces2:"ErrorInterface: OtherMethod SimpleInterfaceMethod"
		SimpleInterface
		SimpleInterface2 // want `embedded interface is not compatible: method "SimpleInterfaceMethod" has mismatches in declared error codes: missing codes: \[interface-1-error] unused codes: \[interface-3-error]`
		OtherSimpleInterface
	}

	EmbeddedSimpleInterfaceOtherFile interface { // want EmbeddedSimpleInterfaceOtherFile:"ErrorInterface: SimpleInterfaceMethod"
		SimpleInterface
	}

	EmbeddedSimpleInterface2OtherFile interface { // want EmbeddedSimpleInterface2OtherFile:"ErrorInterface: SimpleInterfaceMethod"
		SimpleInterface // want `embedded interface is not compatible: method "SimpleInterfaceMethod" has mismatches in declared error codes: missing codes: \[interface-3-error] unused codes: \[interface-2-error]`

		// Errors:
		//
		//    - interface-1-error -- could potentially be returned
		//    - interface-3-error -- could potentially be returned
		SimpleInterfaceMethod() error // want SimpleInterfaceMethod:"ErrorCodes: interface-1-error interface-3-error"
	}
)
