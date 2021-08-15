
complexity tier
---------------

So, how complex is the AST walker, anyway?
Does it need symbolic execution?  :O

I don't think so.
I think we can actually use remarkably simple heuristics and be totally fine.
I think we can just look for any of the errtag functions or structs that receive a tag string...
and just add every string seen in those to the list of errors this function can return.

We do probably need to walk the flow control graph a little to make sure no error values are being returned that _haven't_ passed through one of the errtag functions.
(But it's not at all necessary to try to emulate about which returned values come from what function, or figure out exactly which tags may have been applied by which line, etc.)

We also should be able to walk into functions and tell what tags they return even if they don't have a doc comment about it.
(This will need recursion limits of some kind too -- but doing it within a package seems fine, for starters.)
(This should cover the user story for when people write constructor functions for their errors, which is something we'll probably also encourage.)

... I think there might be a much easier way to deal with some of this: if returned values must be `errtag.Error` rather than `error`, did we just get the golang complier to do all the checks for us?  We did, right?
This would not be a choice to make lightly, because a lot of golang stuff might see this as "weird".  The `error` type gets defacto blessed in a surprising number of ways, some of which are difficult to imagine in advance.
(For example: will other tools like errcheck work if you don't have literally `error` in your returns?  I assume it will not.  It's an ecosystem thing, not just a typesystem thing, at this point.)
Using a type for this would also require an interface, not just a struct, due to the typed-nil problem and the very very (nearly 100%) practical likelihood of encountering it.  (Not a problem, just, know this.)
... No.  This is **not** sufficient.  If we are to be able to have the set of errors returned by a function tree ever become shorter in the higher levels, then we must be able to trace that some errors returned from deeper calls (already tagged!) are known to be converted to another (probably smaller) set.


unsolvables
-----------

### untrackable tag assignment

If we have an interface, then...
we always can have someone set tags in a way we can't track.
That means being able to do naughty things that are not easily statically analyzed.

This is also the case if we have any exported fields.
(This can maybe be avoided by implementing JSONUNmarshal, etc, but... if we want to be codec agnostic, ugh.)

... heh, I suppose we need some attestational wrapper conventions for anything we're deserializing, too, speaking of.

We either have to choose not to have an interface, or, just accept this and ask people to play nice.

### unwrapping cause puts leaves you with any error tag

Sort of fine, honestly.  If that's a problem, well, do you some routing that makes your expectations explicit again.



dismissed ideas
---------------

### making it easy to create custom golang types that match the interface

```
// What do you want your own error types for?  Is it just so you can have native type switches too? // ... no, that would be... bad, ecosystemically.
// Is it so you can have constructors?  // ... no, that's dumb, just do that with functions, much easier.
type MyError struct {
	errtag.ErrorStruct
}
```

### factory functions in errtag itself

```
func DefineError(tag string, msgFormat string) func(details map[string]string) error {
	panic("maybe not a good idea.  can't really make functions that return typed constructors in golang.")
	// An optional extension feature doing a tiny bit of codegen could be useful, though.
}
```
