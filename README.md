go-ree -- **R**e-**E**xamining **E**rrors
=========================================

Error handling is a complex business.
The complexity is high in any program;
it also grows rapidly when programs are networked,
and a user's action may have triggered errors from any one of a series of dependent programs...
which may have been written in multiple languages, etc, etc.

`ree` is:
1. a short language-agnostic spec for error tags -- and recommendations about how to compose them.
2. a slightly bigger (but still short) language-agnostic spec for error serialization which should be reasonably useful from any program, and parsable and chainable by any program.
3. a static analysis tool, for golang, to support you in writing code that follows the `ree` conventions.

This repo is concerned with golang code, and golang tooling,
but the concepts and the error serialization form are language agnostic.


Status
------

This repo contains both a spec about error values and their serializablity,
and a static analysis tool you can use to make sure your program propagates error values,
together with their "codes" (as described below by the spec) correctly.

Both the spec and the static analysis tool are in **beta**.
You can use them!  Please do!
The authors are beginning to dogfood them too.

See [README_recommendedUsage.md](README_recommendedUsage.md) for more information
on how we suggest getting started.


Tagged Errors: a Proposal
-------------------------

The design is derived from a few concepts:

- 0. Errors are Values.
- 1. It is important to be able to handle errors programmatically, and exhaustively: for this we will need errors to have some kind of "code" or "tag".
- 2. It is important for errors to be serializable, and deserializable.  For this reason, we will specify a minimal but standard structure.
- 3. It is important for the error "code" to be textual, so that it can be serialized, and so that programmers can make them rich and unique (as opposed to ints, which would be prone to colision).

This is all a fancy way of saying errors should have a string constant that says what kind of error they are.

If you're buying in so far, let's also continue with a few more design concepts about how we should serialize errors.


Serializable Rich Errors: a Proposal
-------------------------------------

Continuing the list above:

- 4. It is important for errors to describe themselves in a human-readable way: they should be able to store a freetext message that's reasonable prose to show to a user.
- 5. It is useful for errors to be able to attach details: key-value freetext pairs are sufficient for this.
- 6. It is too much to expect that string templating will be available in the same way in every language, so message text is free to repeat parts of the details entries.  (The message is for humans; the details entries are for machines.)
- 7. It is too much to expect that anything at all can be standardized except codes.  For example, it is not reasonable to expect stack traces can be standardized, or even that they are always desirable.  If data such as this are present, they should be placed in a detail entry as freetext, like anything else -- not blessed.
- 8. It is *not* advisable to use the concept of inheritance hierarchies for errors.  Pattern matching and re-tagging are better approaches for holistic reasons, and critically, are viable to reason about locally.

These concepts bring us to a serial format, which wil be described in the [Schema](#schema) and [Serializing](#serializing) sections;
a convention for printing human-readable messages, which will be described in the [Printing](#printing) section;
and, ultimately, will even allow us to build tools for static analysis to check programs for good and complete error handling.


### Schema

```
type Error struct {
	code    String
	message optional String
	details optional {String:String}
	cause	optional [Error]
}
```

### Serializing

Routable Errors are serialable as JSON.
(They're also serializable... however else you like.  It's just strings and maps and lists!)

The following are all examples of Routable Errors expressed in JSON:

```json
{"code":"foobar-error"}
```

```json
{"code":"foobar-error", "message": "this is prose describing the problem"}
```

```json
{"code":"foobar-error-gotzonked", "message": "help, I've been zonked during frobnozing!", "details":{"during":"frobnoz"}}
```

```json
{"code":"foobar-error", "cause":[{"code": "subsys-error"}]}
```


### Printing

Errors must be able to be printed in a human-readable way.
(Printing JSON is fine and dandy, and is certainly _complete_,
but can be excessively verbose and somewhat user-hostile,
so printing JSON shouldn't be our only option.)

The recommended format is:

- if only `code` is present: `{{code}}`.
- if `message` is present: `{{code}}: {{message}}`.
- if `cause` is present, and there's one of them: `{{code}}: {{cause}}` (apply the printing function recursively).
- if `message` and `cause` are both present: `{{code}}: {{message}}: {{cause}}`
- if multiple values are in `cause`: this is undefined.  Consider emitting a list of codes, or just prose about details being elided.  Avoid overwhelming the user (or consider just showing JSON at this point).

Why like this?

- the code is always present, and is the clearest thing (think: end users should definitely be copy the error code into search queries!), so it absolutely must be visible.
- it is not necessary to say "error " as a prefix, because codes will usually say that word already (and it would create stuttering when printing cause chains).
- the message is already prose: there's nothing more that needs to be done with it.
- other errors in the cause can just repeat these patterns.
- the details map is not important to print, because the message prose should already contain the same information, contextualized.
- remember: we don't need to print something that's parsable here, because we do still just have JSON for when we want that!


### Recommended Conventions for Codes

Error codes should be reasonably human readable, reasonably unique,
and should avoid whitespace or other characters that could make them require quoting or escaping.

We recomend that conventional error codes should be defined as all lowercase,
and be composed of several hunks of ASCII text (`[0-9a-zA-Z]`) which are joined by dash characters ("`-`").

The first hunk should be a package name or application name,
or other reasonably unique and contextually relevant string.

It is typical, but not required, to put the word "error" as a second hunk,
especially if the other hunks in the code would not make it clear that string is describing a problematic situation.

The last hunk should describe the specific error.

More intermediate hunks can exist as desired.

This concept of "hunks" is meant to to encourage uniqueness,
while also acknowledging typical logical grouping patterns that are likely to emerge,
but are recommendations only.

Here are some example error code strings which follow the convention:

```
libzow-error-needconfig
libzow-error-frobnozfailed
wowapp-error-unknown
wowapp-error-subsys-bonked
wowapp-error-subsys-zonked
wowapp-error-othersys-storagecorrupted
```


### Code vs Message vs Details

The **code** is supposed to be short clearly a single term.

The **message** is supposed to be prose, ready to show to a human user.

The **details** are freetext, but as a key-value map, can be inspected more individually.
Most of the details should be repeated in the message already, if they're important to a human user.

Libraries for implementing ree patterns will typically have some sort of wrapper functions
for developers to define a function that sets an error code,
accepts some known detail parameters that are expected for this error code,
and templates the details into a message, all at once.
However, it's important to note that all this is done in a library _where the error is created_.
None of this responsibility or complexity is foisted on any other application that wants to parse the error
or display the error to a human.


### Cause is a list

Usually, the `cause`, if present, is a list with one element.

A longer list is allowed, but rarely used.

If you have an API that truly needs to report many errors, you can use the cause list,
but beware that if more than one cause is present,
while all ree libraries should _handle_ that data without loss,
when it comes to printing, many ree libraries may choose not to print all of the causes anyway,
for fear of overloading a human reader.


### When not to use Ree

You can probably pretty much always use `ree` conventions.

Sometimes in high-performance internal APIs (e.g., think: something where you're returning `io.EOF`, etc),
you will want to skip the message component or having any "details" maps.
That's fine.



Conventions in Golang
---------------------

Packages are encouraged to make a list of constants of their error codes, like this:

```
package foo
const (
	ErrBar = "foo-error-bar"
	ErrZaz = "foo-error-zaz"
)
```

Packages are encouraged to make a habit of returning only the error codes from their own package from their functions.
It is reasonable and should be normal to take any of the errors returned by functions you call outside your own package,
and immediately route them into errors which use the codes from your own package.
(Doing this means the enumeration of error codes that any given function returns should remain reasonably short,
because it encourages examining errors and handling them close to where they occur whenever possible,
or at least understanding and categorizing how they should affect a user of the package.
Doing this habitually and ecosystemically should mean every package has a relatively limited and clear number of error codes to return,
and overall make logical responses to error codes easier to build because there's only so much ever needing to be handled in one place.)


Analysis tools in Golang
------------------------

See the `analysis` folder, which contains a distinct go module, and tooling.

This tool will perform static analysis on golang code,
and verify that errors follow the ree convention,
that functions document which ree error codes they can return,
and that those documents about error codes are truthful.

More mechanically speaking:

1. Look for functions which have a comment block which describes error codes;
2. For such functions, statically analyze them (and any other functions they call) to see where data in errors returned actually originates from;
3. Find the ree "code" in those error value origin sites;
4. Ultimately, check that the error codes claimed in the docs are actually true.

(The model used is one of simple tainting, but this is sufficient to reason about code that is reasonably well-structured.)

This tooling should faciliate a conversation between the programmer and the analysis tool:
the programmer writes a claim in their function docs,
and then the analyzer checks it.
If the programmer fails to describe any errors that the code actually does return,
the analyzer will prompt them to fix this.
If the programmer finds that they're having to document too many error codes,
then they will be encouraged to refactor their code until the error handling becomes reasonable.


The code in this package
------------------------

_You can actually ignore it._

The code in this package implements the ree conventions.

But it's not special.  You could write similar code in your own repo,
and the ree analyzer tool will treat it exactly the same way.


License
-------

SPDX-License-Identifier: Apache-2.0 OR MIT
