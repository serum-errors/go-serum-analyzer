package ree

import "fmt"

type Error interface {
	Code() string
	Message() string
	Details() map[string]string // may optionally be present and contain itemized details.  The same content should be present in the message, preformatted.
	Cause() error
	Error() string // compliance with `error`.  For our errors, this will return "{tag}: {message}[: {cause}]".
}

// ErrorStruct is an exported struct so you can use e.g. json marshal and unmarshal on it easily.
type ErrorStruct struct {
	TheCode    string            `json:"code"`
	TheMessage string            `json:"msg,omitempty"`
	TheDetails map[string]string `json:"details,omitempty"`
	TheCause   *ErrorStruct      `json:"cause,omitempty"` // note the concrete type -- necessary, so unmarshal can work recursively.
	// TODO: consider changing TheCause to a list.  (`Cause()` can return an errorList type, if it likes to remain simple for the common case.)
}

func (e *ErrorStruct) Code() string               { return e.TheCode }
func (e *ErrorStruct) Message() string            { return e.TheMessage }
func (e *ErrorStruct) Details() map[string]string { return e.TheDetails }
func (e *ErrorStruct) Cause() error               { return e.TheCause }
func (e *ErrorStruct) Error() string {
	switch {
	case e.TheCause == nil && e.TheMessage == "":
		return e.TheCode
	case e.TheCause == nil:
		return fmt.Sprintf("%s: %s", e.TheCode, e.TheMessage)
	case e.TheMessage == "":
		return fmt.Sprintf("%s: %s", e.TheCode, e.TheCause)
	}
	return fmt.Sprintf("%s: %s: %s", e.TheCode, e.TheMessage, e.TheCause)
}

// tbd: how this should interact with `errors.Is`, etc.
