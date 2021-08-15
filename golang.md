
- I want people to be able to define their own error interface and use it as a marker for their function return types.
	- This lets the go compiler do quick checks that are less disruptive to developer flow than running the linter, and helps.
	- But it should also be optional, and if someone wants to stick with plain `error` all the way, that should not be a problem either.

- It would be nice if the linter could run on code that doesn't import any library at all.
	- ... but I don't know how to do that.  We need landmarks for what code assignment looks like, and what switching on codes looks like.
		- maybe this is actually possible: look for interface matches, and look at the body of the `Code() string` method...
			- if that returns constants, we're done and happy.  (and someone did a strong form of typed errors; wow, bold.)
			- if that returns a single struct field, we can still look at that, and now we know what type and field name is has the concept of "code" here, and look for assignments to that or struct initializers.
				- does not save you from deserialization (as it probably uses reflection), but, that is not a new issue.

- I had a first draft that had very explicit "router" functions, but I think that may have been a mistake.
	- One ends up writing a DSL for this...
		- ... and then one ends up debugging the performance of the DSL...
			- ... this way does not lie joy, I fear.

- There are several kinds of startpoint:
	- Assigning a variable from a new initializer.  (Duh, and the easy case.)
	- Assigning a variable from another variable on the stack.  (Just follow it.)
	- Assigning a variable from a function result.  (Recurse into this.)
		- ... if it's statically obvious!  If it's an interface, well, that's another matter.
	- Assigning a variable from a function result that's an interface call.  (This'll be fun!)
		- I think interfaces just have to say what they promise, and that's the end of it.
			- Probably a good thing anyway, isn't it?  An interface that doesn't declare its error returns is a pretty lame interface to code against.
		- Has an interesting implication though: we're gonna want to analyse what interface any publicly exported structs implement, and then check those contracts match too?
			- Might demand that people use specific conventions to declaring the interface relationship explicitly, because otherwise this is a serious problem.
	- Assigning a variable from some other heap variable.  (This'll be fun!)
		- I have no idea how we'll do this.  Probably cop out and say that's unanalysable and make the developer explicitly re-tag the thing.  (I think this is uncommon enough that this is probably fine.)
		- Can try to improve on this over time, perhaps, while skipping it in the MVP.
	- Someone assigns the code field of an existing thing.  (This'll be fun!)
		- I think it would be nice to just say "don't do that" by fiat, but I'm not sure if that's possible without conflicting with other golang usability goals.
	- Creating something by deserialization.  (This'll be fun!  Double fun because it's probably done by reflection.)
		- Means: we'll just have to handle the situation of an error value appearing that we can't explain yet, and require explicit user tagging there (and have to have faith in it!).
		- We full-on have the runtime excursion problem here.  (It's never going away.)  The best thing we can promise is that we can at least, at runtime, keep returning stuff without losing data.
			- I wonder if this is going to be a strong argument for needing our own router function DSLs, after all.  Hand-written switches are gonna make mistakes on this _all_ the time.
	- Assigning from within a closure.  (Uffdah.)
		- This is... maybe it's fine.  Since we're starting our tracing only from return values, we don't have to worry about *when* closures are being called, and can continue to use the logic about "any tags seen are possible".
	- Assigning a variable from a recover call!  (Oh my!)
		- F a s c i n a t i n g.  If we make some rules (like only checking up through one package), we might actually be able to reason about this fairly usefully?
		- Not mvp though, alright?
		- Fallback: treat this the same as an interface function call, and require explicit redeclaration.
		- this mostly probably still just accumulates things, though, is the trouble.  (If you already used your return error value as a junkdrawer, this won't fix the taint explosion that you already had.)

- Random thought: `defer ree.FinalWord(&err, func(err error) error { /.../ })` could be used as an option to handle taint explosion.
	- the pure function part (and then, _our_ code ensuring there's ultimately an assign at the end) lets us fix the taint explosion.
		- would also need to check that this is the first (or last, depending on your perspective) defer for this function.
	- not needed in mvp.  not needed at all.  just an interesting option.

---

there are two separate problems we're solving:

- the first is getting docs on functions to enumerate their error's codes.
	- this can be highly recursive, and only needs to trace assignments.
	- and yes, it's going to have the potential for taint blowup if you ever return one of your variables that's been overwritten repeatedly.  (but so _don't_.)
	- _this covers a lot of ground_.  it's enough, by itself, to give you the tools not to have explosions of error codes (or uncoded terrors).
		- the one thing it doesn't help you with is, when you're writing the logic to re-code any errors from child calls... you probably want to write a switch statement or something.
			- and the default case?  you'd like to not need one, statically, right?
- the second is being able to check that your functions have exhaustively handled errors from their child calls.
	- _this is a whole separate endeavour._  the first endeavour provides value _on its own_.
		- you can always do a panic as the default path in your error handling, and that solves your issues.  (you just never want to have a fallback that's returning the junk error var, because taint explosion.)
		- or do an "*-unknown" error code.  that's legitimate.
	- mind that there will _still_ be situations where you want to have an "*-unknown" error code.
		- remember, this is legitimately true any time you try to process an error code that comes from a deserialization of a message from another program(/compilation-unit).
			- you'll want to wrap whatever-it-is and pass it up truthfully.
				- a higher level of your program will eventually probably shift it into a vaguer error code that blames your whole subsystem, but will keep the cause chain intact, and that's... a good thing.
	- we'll probably have to make this opt-in per site.  I don't think it'll be accepted if it's sounding off by default.
	- we *can* make a default case with a panic be what you write, and what we check, then.
	- we'll only be able to recognize so many patterns.  (Mostly focusing on switch cases, probably.)
		- the first pattern will be switches on literally the code string value.  no prefixes or suffixes or wildcards, nothing fancy at all: literals.
			- it will also probably require that the error switch (and the unconditional type cast for the interface, and the `Code()` unboxing), comes literally after (in the AST) an assignment to an error value.
				- this avoids the taint explosion problem, and is the only way I can really imagine this working in syntactically common golang.
				- one could also make a convention of `x, err = func(){ x, err := foo(); switch err.Code{ /.../ }}` to get a new scope and dodge the taint problem, but... considering that still needs the Three Cursed Lines afterwards, it's unclear if this is a useful idea.
		- in general it has to be something that clearly checks all the options at once, because we don't want to get into symbolic execution territory (still).  (switch statements accomplish this!)
	- we can add more patterns over time!
		- maybe the router function idea will actually be good.
		- means we can add fancy patterns that encourage you to use the library, without making them mandatory.  (and if you don't use them, well, you just have verbosity to pay; it's nice to leave that choice on the table.)
	- you'll also want the good sense and judgement not to try to do an exhaustiveness check with a panic as the default case if the function you're calling is in another module (i.e. versioned separately).
		- or maybe you do; i'm not your boss.  it's shakey ground to try to colonize, though.
