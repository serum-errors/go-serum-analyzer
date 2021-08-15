package main

import (
	"fmt"
	"os"

	ree "github.com/warpfork/go-rerr"
)

func main() {
	Coolfunc()
}

// Coolfunc does some cool stuff, and uses ree errors.
//
// Errors:
//
//   - cool-nozzle-bork -- when the nozzle is bork, silly.
//   - cool-sozzle-fizz -- when there's too much fizz in the sozzle.
//   - razzle-* -- if the razzle subsystem errors, we return that without modification.
//
// The above comment block (starting with "^// Errors:$", and through the bullet points on subsequent lines) is machine-parsible.
// The ree system has a linter tool that validates both the comment format (easy),
// and that the error tags listed in the doc comment actually matches the set of error tags
// that the function could possible produce (by analyzing the function's AST).
//
func Coolfunc() error {
	// Any error can be "tagged".
	// The simplest way to do this is the `ree.Tag` function,
	// which takes an error pointer, and if it contains any vaue,
	// will replace it with a new ree.Error which wraps the old error and attaches the given tag string.
	// (This doesn't let you attach any message or details; other functions are needed for that.)
	err := ThingThatErrorsUnknownly()
	ree.Tag(&err, "cool-nozzle-bork")
	if err != nil {
		return err
	}

	// Sometimes you need to decide how to tag an error by looking at some properties of it.
	// If the the tag is affected,
	err = OtherThingThatErrorsUnknownly()
	ree.TagByRouting(&err,
		ree.Route{"cool-nozzle-bork", os.IsExist},
		ree.Route{"cool-sozzle-fizz", ree.MatchTag("cool-sozzle-fizz")},
		ree.Route{"razzle-*", nil},   // A wildcard plus no predicate means passthrough errors with tags matching the wildcard.
		ree.Route{"", os.IsNotExist}, // Empty string means ignore it.  Use with caution.
		// The unavoidable final fallback is, if none of these routes match... a runtime panic.
		// (Or just returning a dumb, untagged error anyway, and potentially causing further chaos up the stack.)
		// However, the point of ree is... if you've used the linter, you can avoid that possibility in advance.
	)
	if err != nil {
		return err
	}

	// Sometimes you already have all ree errors, and can just map them.
	// ... actually, that's not even very interesting.  We already did that with the routes and predicates example above.

	// TODO: example where message and details are used.
	// The current "route" examples aren't powerful enough for that, neither are the Tag functions.

	// TODO: benchmark if whatever DSL this results in is actually acceptable overhead.  If it allocates a slice every time... yikes.
	// Alternative: can we make a convention that works from "you make a switch statement, and each case must contain a recognizable tagging function?
	// This might get a bit tricky, if we want it to recognize custom "constructors".  Is this just ensuring a logical block contains no other recognized errors?  Maybe that's just plain code reuse of what we're already doing.

	return nil
}

func ThingThatErrorsUnknownly() error {
	return fmt.Errorf("unknown nonsense")
}
func OtherThingThatErrorsUnknownly() error {
	return fmt.Errorf("other unknown nonsense")
}
