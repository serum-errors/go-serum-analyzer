go-serum-analyzer
=================

go-serum-analyzer is a static analysis tool for
the use of [Serum](https://github.com/serum-errors/serum-spec) errors in Golang.

The analyzer verifies that Serum errors are propagated, handled, and documented accurately.

Keep scrolling down for a [quick example](#quick-example), or,
jump to the [USERGUIDE](USERGUIDE.md) for lots and lots of examples!


In Brief
--------

- Analysis uses a taint model: all possible codes that a function can return are determined; then the analyzer requires that the documentation of the function states the same set, or emits warnings about the difference.
- Analysis is mostly automatic and you control it by structuring your code flow to handle errors.  In cases where your code is more complex than the analyzer can understand, you can manually declare changes to the expected error set on return sites, if necessary.
- Declarations of the errors a function returns are made in documentation on the function, in a simple and readable convention.
- Works with any `error` type that has a `Code() string` method -- your project doesn't need any special dependencies to use this tool!

This tooling should faciliate a conversation between the programmer and the analysis tool:
the programmer writes a claim in their function docs,
and then the analyzer checks it.
If the programmer fails to describe any errors that the code actually does return,
the analyzer will prompt them to fix this.
If the programmer finds that they're having to document too many error codes,
then they will be encouraged to refactor their code until the error handling becomes reasonable.


Quick Example
-------------

Here's a quick, very simplifed example of how code and structured comments are used
to communicate with the analyzer tool and produce more reliable code:

```go
// This is a function that returns a Serum-style error.
// The doc block below is validated by the analyzer tool!
//
// Errors:
//
//    - examples-error-foobar -- if things are foobar'd.
//    - examples-error-discombobulated -- if caller should recombobulate.
func DoSomething(param interface{}) error {
    if param == nil {
        return &Error{"examples-error-discombobulated"}
    }
	return &Error{"examples-error-foobar")
}

// This is a function that calls another function that returns a Serum-style error.
// The analyzer will make sure that we document all the errors that
// can emerge from this function, and the functions it called.
//
// Errors:
//
//    - examples-error-foobar -- comes from DoSomething.
//    - examples-error-discombobulated -- also comes from DoSomething.
func DoMoreThings() error {
	return DoSomething(nil)
}
```

If you forgot to list some errors?  The analyzer will warn you.

If you listed errors that aren't really there?  The analyzer will warn you.

If you add more errors to a function you wrote earlier, but forget to update either the docs or the handling logic in functions that use it?  _The analyzer will warn you!_

Of course, this is just the beginning, and many styles of usage are supported.

See the [USERGUIDE](USERGUIDE.md) for more!



Adopting
--------

You should be able to start using the Serum error conventions, and this analyzer too, **incrementally** -- _anywhere_ in your stack.


### Mechanically, how is it incremental?

- The analysis tool uses the standard golang analysis libraries.
	- This means you can start using it wherever you'd use `go vet` or other analysis tools, and treat it the same way.
- There is no concrete error type you have to use in your code -- you just have to implement both `error` and `interface { Code() string }`, and the analyzer tool will detect it and figure it out!
  	- This means you can start adopting Serum standards and using go-serum-analyzer without adding even a _single_ import to your package tree.  No dependency hell!  No real risk at all.
- There is no required change to error interfaces you advertise publicly.  All your methods can still return the usual golang `error` interface.
	- This means your public APIs don't change at all if you start adopting Serum standards and using go-serum-analyzer.  No breakage; not even visual surprises.  And any other tooling that looks for `error` types can continue to work the same as always, too.
- You can use go-serum-analyzer on one package at a time.
	- This means you don't have to adopt it in a whole project at once!  You can pick where to begin, and increase your usage slowly if it works well for you.
- The go-serum-analyzer treats packages as natural boundaries.
	- This means it's very natural to adopt the Serum conventions within one package, and massage any not-yet-Serum-style data at the site it calls out to other packages.  In fact, that's the normal way to work even when Serum conventions are fully established!  So the incremental adoption journey is actually the exact same as the full adoption final results!
- Nothing is checked if a function doesn't contain the "Error:" block in the docs -- so you can start using it on whichever functions you prioritize, and adopt it on others at a rate of your choosing.

Overall: it should be very easy to start using the Serum error convention without major up-front hurdles.
It should also be very easy to start using this analzyer tool incrementally, and get good results.


### Any recommendations about where to start using the Serum conventions and the analyzer?

Anywhere is fair game!

It's usually most fun to start in one place and spread out from there.

So far, we've found it usually feels most rewarding to start at either the top or the bottom.
Here, "top" means somewhere by a CLI or API,
while "bottom" means somewhere packages that are farthest to the leaves of your dependency graph.

Working from the top down is rewarding because it lets you start by enumerating errors in your public API first
(which is a good practice anyway!), and then applying the analyzer tool and fixing issues it reports
will make sure that your program is handling errors gracefully and making them eventually fit into the public API you defined.

Working from the bottom up isn't necessarily as rewarding, but it's usually easier: because you usually have less diversity of incoming error causes,
so it's relatively easy to inspect them and decide case-by-case what to do with them.
Then, that rolls forward into helping you decide what errors still need to be reported by your package.
It's still plenty of work, but the less you have to look deeply and recursively into code you're calling, the more tractable it is.

Working from the middle out is possible, but tends to be a bit trickier.
It's not that the analyzer tool works any differently!
It's just that it can be tricky to decide if errors should be filtered differently
in the package you're working on, or in the package below that it's calling, or the package above that the values are reported to.
Usually those same questions are clearer to answer when either at the top or bottom of a stack,
because there, you can fight on one front at a time, so to speak, which is easier.
At the same time, you're not getting user-facing improvements yet, so it's less instantly rewarding.
It can still be worth it to start this way in some codebases, though;
perhaps you need to start increasing code quality _somewhere_, and the easiest place to patch happens to be in the middle; so be it!

Use your best judgement!
No matter where you start, it should be possible to make things better incrementally, so just pick a place and dig in!


Success Stories
---------------

One of the first projects to deploy the go-serum-analyzer "in the wild" was [Warpforge](https://github.com/warpfork/warpforge/).

In very little time, we found the analyzer was replacing the equivalent of **thousands** of manual checks, and greatly increasing the quality of development in the packages that adopted it.

**By the numbers**, we found:

- 983 errorcode-sites verified by automation (meaning: error codes per function, in those functions that are explicitly annotated).
- 59 functions were manually annotated to reach this result.
- 174 functions in total were covered implicitly, since analysis covers all functions called by an annotated function.
- (That means annotating 33% of functions was enough to attain the full effect!)

**Qualitatively**, one contributor notes:

> We were able start getting error codes and linting in place in one pass, then do a second pass to clean them up (with the advantage of now having the linter).
> Having tools made it less daunting to improve iteratively!
>
> Trying to do the same kind of iterative improvements _without_ tool support has never seemed to work out well for us in golang before, because there's too much mental "cache invalidation" and manual review required, making it intractable -- so this was a big deal.
>
> Our package boundaries and code organization makes more sense now.  Having to think explicitly about how many error codes a function returned made us want to simplify it.  That turned into a good influence!
>
> We're also much happier about our docs.

Another the contributor had this to say about it:

> i don't like that this linter has to exist
>
> but i do like that it does exist

((_Translation:_ it would be even better if this was a core tool!  But having it as a linter phase is better than not having it.))

In terms of **observed adoption rate**:
The Serum analyzer tool was added to CI in this project within the same day that Serum annotations were first introduced in some functions of some packages.
Strict mode -- meaning public functions _without_ serum annotations are flagged as _errors_ -- was adopted only _three days_ later!

Overall, we consider this a **rousing success**.
The Serum Analyzer proved both useful and smoothly incrementally adoptable,
and positively influenced the code quality _and_ the developer satisfaction within the adopting project.



License
-------

SPDX-License-Identifier: Apache-2.0 OR MIT
