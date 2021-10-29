package analysis

import (
	"fmt"
	"strings"
)

type state interface {
	step(stateMachine *findErrorDocsSM, line string) error
}

// findErrorDocsSM is a state machine (SM) to find error codes in the provided doc string.
// It does so by looking at the doc comments on a function,
// then trying to parse out error code declarations that we can recognize,
// and finally returning the error code strings from that.
//
// The declaration format is:
//   - strip a leading "^//" if present.
//   - strip any more leading whitespace.
//   - a line that is exactly "Errors:" starts a declaration block.
//   - exactly one blank line must follow, or it's a bad format.
//   - the next line must match "^- (.*) --", and the captured group is an error code.
//     - note that this is after leading whitespace strip. (probably you should indent these, for readability.)
//     - for simplier parsing, any line that starts with "- " will be slurped,
//       and we'll consider it an error if the rest of the pattern doesn't follow.
//     - the capture group can be stripped for whitespace again. (perhaps the author wanted to align things.)
//     - the error code has to be valid, which means it has to match against: "^[a-zA-Z][a-zA-Z0-9\-]*[a-zA-Z0-9]$" or "^[a-zA-Z]$"
//   - for error constructors lines like "^- param: (.*) --" are allowed.
//     - the captured group has to be a parameter of type string
//   - this may repeat. if lines do not start that that pattern, they are skipped.
//      - note that the same code may appear multiple times. this is acceptable, and should be deduplicated.
//   - when there's another fully blank line, the parse is ended.
//   - instead of the standard decleration block it is also permitted to declare that a function does not return codes.
//      - such a line must match "Errors: none(.*)".
//      - this allows documentation; for example: "Errors: none -- this method only returns error to comply with the foobar interface."
// This format happens to be amenable to letting you write the closest thing godocs have to a list.
// (You should probably indent things "enough" to make that render right, but we're not checking that here right now.)
//
// If there are no error declarations, (nil, nil) is returned.
// If there's what looks like an error declaration, but funny looking, an error is returned.
type findErrorDocsSM struct {
	seen      CodeSet
	state     state
	noCodesOk bool
	param     string
}

// run runs the state machine to find error codes in the provided doc string.
//
// The method returns a set of found codes,
// a bool which is true if the function declared "Errors: none",
// an error in case of invalid doc strings or nil otherwise.
func (sm findErrorDocsSM) run(doc string) (CodeSet, string, bool, error) {
	sm.seen = CodeSet{}
	sm.state = stateInit{}
	sm.noCodesOk = false
	sm.param = ""

	for _, line := range strings.Split(doc, "\n") {
		line := strings.TrimSpace(line)
		err := sm.state.step(&sm, line)
		if err != nil {
			return nil, "", false, err
		}
	}

	return sm.seen, sm.param, sm.noCodesOk, nil
}

type (
	stateInit          struct{}
	stateNeedBlankLine struct{}
	stateParsing       struct{}
	stateDone          struct{}
)

func (stateInit) step(sm *findErrorDocsSM, line string) error {
	if line == "Errors:" {
		sm.state = stateNeedBlankLine{}
	} else if strings.HasPrefix(line, "Errors: none") {
		sm.noCodesOk = true
		sm.state = stateDone{}
	}
	return nil
}

func (stateNeedBlankLine) step(sm *findErrorDocsSM, line string) error {
	if line == "" {
		sm.state = stateParsing{}
		return nil
	} else {
		return fmt.Errorf("need a blank line after the 'Errors:' block indicator")
	}
}

func (stateParsing) step(sm *findErrorDocsSM, line string) error {
	switch {
	case line == "":
		sm.state = stateDone{}
	case strings.HasPrefix(line, "Errors:"):
		return fmt.Errorf("repeated 'Errors:' block indicator")
	case strings.HasPrefix(line, "- "):
		end := strings.Index(line, " --")
		if end == -1 {
			return fmt.Errorf("mid block, a line leading with '- ' didnt contain a '--' to mark the end of the code name")
		}

		if end < 2 {
			return fmt.Errorf("an error code can't be purely whitespace")
		}
		code := line[2:end]
		code = strings.TrimSpace(code)
		if code == "" {
			return fmt.Errorf("an error code can't be purely whitespace")
		}

		if strings.HasPrefix(code, "param:") {
			param := code[len("param:"):]
			param = strings.TrimSpace(param)
			switch {
			case param == "":
				return fmt.Errorf("an error code parameter can't be purely whitespace")
			case sm.param != "":
				return fmt.Errorf("cannot define more than one error code parameter (found multiple 'param:' inidicators)")
			default:
				sm.param = param
				return nil
			}
		}

		if !isErrorCodeValid(code) {
			return fmt.Errorf("declared error code has invalid format: should match [a-zA-Z][a-zA-Z0-9\\-]*[a-zA-Z0-9]")
		}

		if _, exists := sm.seen[code]; !exists {
			sm.seen[code] = struct{}{}
		}
	}
	return nil
}

func (stateDone) step(sm *findErrorDocsSM, line string) error {
	if strings.HasPrefix(line, "Errors:") {
		return fmt.Errorf("repeated 'Errors:' block indicator")
	}
	return nil
}
