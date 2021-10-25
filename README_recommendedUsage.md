Recommended Usage
=================

The static analysis tool is now available and you should try it.

You should be able to start using the go-ree conventions **incrementally**, _anywhere_ in your stack.
However, some scenarios may be easier and feel more immediately rewarding than others,
so we'll discuss that here.

### mechanically, how is it incremental?

- The analysis tool uses the standard golang analysis libraries.
	- This means you can start using it wherever you'd use `go vet` or other analysis tools, and treat it the same way.
- There is no concrete error type you have to use in your code -- you just have to implement both `error` and `interface { Code() string }`, and the analyzer tool will detect it and figure it out!
  	- This means you can start adopting go-ree standards without adding even a _single_ import to your package tree.  No dependency hell!  No real risk at all.
- There is no required change to error interfaces you advertise publicly.  All your methods can still return the usual golang `error` interface.
	- This means your public APIs don't change at all if you start adopting go-ree standards.  No breakage; not even visual surprises.  And any other tooling that looks for `error` types can continue to work the same as always, too.
- You can use the ree analzyer on one package at a time.
	- This means you don't have to adopt it in a whole project at once!  You can pick where to begin, and increase your usage slowly if it works well for you.
- The ree analyzer treats packages as natural boundaries.
	- This means it's very natural to adopt the ree conventions within one package, and massage any not-yet-ree-style data at the site it calls out to other packages.  In fact, that's the normal way to work even when ree conventions are fully established!  So the incremental adoption journey is actually the exact same as the full adoption final results!

Overall: it should be very easy to start using the ree convention without major up-front hurdles.
It should also be very easy to start using the ree analzyer tool incrementally, and get good results.

### where should I start using the ree conventions?

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
