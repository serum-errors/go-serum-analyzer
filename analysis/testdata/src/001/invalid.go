package main

func ErrorNotLast() (error, int) { // want "error should be returned as the last argument"
	return &Error{"hello-error"}, 0
}
