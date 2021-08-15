package rerr

import "strings"

func MatchTag(tag string) func(error) bool {
	return func(e error) bool {
		e2, ok := e.(Error)
		if !ok {
			return false
		}
		// future: add support for exactly one '*'.
		return e2.Tag() == tag
	}
}

func MatchErrorWithMessageFragment(frag string) func(error) bool {
	return func(e error) bool {
		return strings.Contains(e.Error(), frag)
	}
}
