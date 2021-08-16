package ree

import "strings"

func MatchCode(code string) func(error) bool {
	return func(e error) bool {
		e2, ok := e.(Error)
		if !ok {
			return false
		}
		// future: add support for exactly one '*'.
		return e2.Code() == code
	}
}

func MatchErrorWithMessageFragment(frag string) func(error) bool {
	return func(e error) bool {
		return strings.Contains(e.Error(), frag)
	}
}
