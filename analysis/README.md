# Error Code Analysis

The analyser in this package checks:

* if error returning functions declare possible error codes.
* if the declared error codes match what is actually returned.

How to run the analyser as stand-alone tool:

* Installing reeverify: `go install ./cmd/reeverify`
* Change Directory to Target Project: `cd <target-path>`
* Execute Analyser: `reeverify <package>`

## Examples

All examples can be found under [testdata/src/examples/](testdata/src/examples/) and they are executed as part of the test suite when executing `go test` inside the current folder.

## Error Type in Examples

The following type will be used in most of the examples:

```go
type Error struct {
    TheCode string
}

func (e *Error) Error() string { return e.TheCode }
func (e *Error) Code() string  { return e.TheCode }
```

`*Error` implements the `error` interface and also is a ree error because `interface { Code() string }` is implemented. Those two conditions are the only ones needed to be considered by the error code analysis.

For more examples of possible or invalid error types see: [testdata/src/errortypes/](testdata/src/errortypes/)

## Declaring and Returning Errors

The following code shows an example application where error codes are declared and errors with the declared codes are returned.

```go
// Add adds a non-nil value into the collection.
//
// Errors:
//
//    - examples-error-invalid-arg        -- if the given argument is nil
//    - examples-error-invalid-collection -- if the given collection is nil or invalid
//    - examples-error-limit-reached      -- if the limit of values in the collection is reached
func (c *Collection) Add(item interface{}) error {
    if item == nil {
        return &Error{"examples-error-invalid-arg"}
    }

    if c == nil || c.limit < 0 {
        return &Error{"examples-error-invalid-collection"}
    }

    if len(c.values) >= c.limit {
        return &Error{"examples-error-limit-reached"}
    }

    c.values = append(c.values, item)
    return nil
}
```

### Error Declaration

Error declarations are done in the comments before an error returning function and have to formatted as follows:

* A line containing exactly `Errors:` starts a declaration block.

* Exactly **one blank line** must follow.

* The next lines are **error code** declarations:

  * Declarations match the format: `- <code> -- <comment>`
  * `<code>` has to be composed of several hunks of ASCII text, which are joined by dash characters ("-"). More precisely `<code>` has to match `^[a-zA-Z][a-zA-Z0-9\-]*[a-zA-Z0-9]$` or `^[a-zA-Z]$`. Find more information on recommended conventions for codes under: [../README.md](../README.mc#recommended-conventions-for-codes)
  * `<comment>` can be any string.
  * Any number of **whitespaces** are allowed before and after every part of a declaration.
    * In the example above whitespaces are used to align the double dashes ("--") after the `<code>` part.

* **Comments** are allowed between error code declarations:
  * Any line that does not start with a dash character ("-") is regarded as a comment.

* The **declaration block ends**, when there's another fully blank line

### Declare No Errors

Alternatively it is allowed to declare that a function returns no errors:

```go
// Read for MockedIO always returns 0.
// Errors: none -- this method only returns error to comply with the interface IO.
func (MockedIO) Read() (byte, error) { return 0, nil }
```

The rules for doing so are as follows:

* The declaration for no errors must match: `Errors: none (.*)`
* This allows for a comment after the declaration. (See example above)
* A function can only have at most one `Errors:` or `Errors: none` declaration (and never both)

### Function Analysis

The analysis tries to find mismatches of declared error codes and actually returned ones. Meaning the tool will complain if:

* error codes are declared, but never returned
* error codes are returned, but not declared

The following examples demonstrate that. The first example declares error codes, but doesn't return any of them.

```go
// Errors:
//
// None of the following errors are actually returned yet.
//    - examples-error-invalid-arg        -- if the given argument is nil
//    - examples-error-invalid-collection -- if the given collection is nil or invalid
//    - examples-error-limit-reached      -- if the limit of values in the collection is reached
func (c *Collection) AddUnused(item interface{}) error {
    panic("not implemented")
}
```

The tool would emit the following output for the example above.

```text
...\testdata\src\examples\02_basic_examples.go:37:1: function "AddUnused" has a mismatch of declared and actual error codes: unused codes: [examples-error-invalid-arg examples-error-invalid-collection examples-error-limit-reached]
```

The second example returns error codes, but does not declare any of them.

```go
// Errors: none -- not actually true, but we want to showcasse missing error codes.
func (c *Collection) AddMissing(item interface{}) error {
    if item == nil {
        return &Error{"examples-error-invalid-arg"}
    }

    if c == nil || c.limit < 0 {
        return &Error{"examples-error-invalid-collection"}
    }

    if len(c.values) >= c.limit {
        return &Error{"examples-error-limit-reached"}
    }

    c.values = append(c.values, item)
    return nil
}
```

Very similar to the first example, the tool would emit the following message for the second example.

```text
...\testdata\src\examples\02_basic_examples.go:45:1: function "AddMissing" has a mismatch of declared and actual error codes: missing codes: [examples-error-invalid-arg examples-error-invalid-collection examples-error-limit-reached]
```

### Alternative Code Styles

We try to support a lot of different programming styles. A previous example could be rewritten to have only a single return statement.

```go
// AddAlt adds a non-nil value into the collection.
//
// Errors:
//
//    - examples-error-invalid-arg        -- if the given argument is nil
//    - examples-error-invalid-collection -- if the given collection is nil or invalid
//    - examples-error-limit-reached      -- if the limit of values in the collection is reached
func (c *Collection) AddAlt(item interface{}) error {
    var err error

    switch {
    case item == nil:
        err = &Error{"examples-error-invalid-arg"}
    case c == nil || c.limit < 0:
        err = &Error{"examples-error-invalid-collection"}
    case len(c.values) >= c.limit:
        err = &Error{"examples-error-limit-reached"}
    default:
        c.values = append(c.values, item)
    }

    return err
}
```

## Limitations

TODO:

* Dead branches cannot be detected
* Error has to be last return
