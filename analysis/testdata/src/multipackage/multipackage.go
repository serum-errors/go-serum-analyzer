package multipackage

import (
	"multipackage/inner1"
	"multipackage/inner2"
)

// RunPackage1 is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned
func RunPackage1() error { // want RunPackage1:"ErrorCodes: hello-error"
	return inner1.ExportedFunc1()
}

// RunPackage2 is a demo function
//
// Errors:
//
//    - hello-error -- is always returned
func RunPackage2() error { // want RunPackage2:"ErrorCodes: hello-error"
	return inner2.ExportedFunc2()
}

// Trap1 is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned
func Trap1() error { // want Trap1:"ErrorCodes: hello-error"
	trap := returnTrap()
	return trap.returnError()
}

// Trap2 is a demo function.
//
// Errors:
//
//    - hello-error -- is always returned
func Trap2() error { // want Trap2:"ErrorCodes: hello-error"
	return returnTrap().returnError()
}

type TrapType struct{}

func (TrapType) returnError() error {
	return nil
}

func returnTrap() TrapType {
	return TrapType{}
}
