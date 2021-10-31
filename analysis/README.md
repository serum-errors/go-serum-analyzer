# Error Code Analysis

The analyser in this package checks:

* if error returning functions declare possible error codes.
* if the declared error codes match what is actually returned.

How to run the analyser as stand-alone tool:

* Installing reeverify: `go install ./cmd/reeverify`
* Change Directory to Target Project: `cd <target-path>`
* Execute Analyser: `reeverify <package>`

## About Examples

All examples can be found under [testdata/src/examples/](testdata/src/examples/) and they are executed as part of the test suite when executing `go test` inside the current folder.

### Error Type in Examples

The following type will be used in most of the examples:

```go
type Error struct {
    TheCode string
}

func (e *Error) Error() string { return e.TheCode }
func (e *Error) Code() string  { return e.TheCode }
```

`*Error` implements the `error` interface and also is a ree error because `interface { Code() string }` is implemented. Those two conditions are the only ones needed to be considered a ree error by the error code analysis.

More detail about error types can be found in a later section of this document.

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

The following two examples demonstrate that. The first example declares error codes, but doesn't return any of them.

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

The second example returns errors, but does not declare any of the returned codes.

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

## Error Types

To be considered a valid ree error a type must implement the following interfaces:

* `error`, which is `interface { Error() string }`
* `interface { Code() string }`

For the `Error` method, there are no limitations to how it may be implemented. We recommend to follow the style laid out in: [../README.md](../README.md).

The `Code` method is used to get the error code from an error type instance. For the analysis tool to be able to statically track error codes, the `Code` method has to follow some conventions:

* return a constant string, or
  * Multiple different constant strings may be returned.
  * The returned strings have to be error codes of a valid format (i.e. has to match `^[a-zA-Z][a-zA-Z0-9\-]*[a-zA-Z0-9]$` or `^[a-zA-Z]$`)
* return a field
  * If a field is returned multiple times, always the same field has to be returned.
  * We call this field the **error code field**.
  * When constructing an error, this field has to be set to a constant string.

There may be a mix of returning the error code field and constatns inside of one `Code` method.

Assignment to error code fields is also restricted to make static analysis possible:

* Assignments to the error code field have to be constant strings.
  * If a constant string is assigned in a method of the error type: that string is added to the possible error codes the error type may return.
  * If the assignment is done in a function returning the modified error: the assigned string is added to the possible error codes this function may return.
* All other assignments are prohibited.

### Examples

We have already seen the following simple example, where a single error code field is returned every time:

```go
type Error struct {
    TheCode string
}

func (e *Error) Error() string { return e.TheCode }
func (e *Error) Code() string  { return e.TheCode }
```

The next example shows how constant string error codes could be returned:

```go
type Error2 string

const (
    errorPrefix     = "examples-error-"
    ErrorEmpty      = errorPrefix + "empty"
    ErrorDisconnect = errorPrefix + "disconnect"
    ErrorUnknown    = errorPrefix + "unknown"
)

func (e Error2) Error() string { return fmt.Sprintf("%s: %s", e.Code(), e) }
func (e Error2) Code() string {
    switch {
    case e == "":
        return ErrorEmpty
    case strings.HasPrefix(string(e), "peer disconnected"):
        return ErrorDisconnect
    default:
        return ErrorUnknown
    }
}
```

Whenever `Error2` is returned anywhere, the analysis adds all three error codes as possible outcomes. So a function with the statement `return Error2("error message")` has to declare the error codes: "examples-error-empty", "examples-error-disconnect", and "examples-error-unknown"

As a final example we show how returning fields and constants can be mixxed:

```go
type Error3 struct {
    flag bool
    code string
}

func (e *Error3) Error() string { return e.code }
func (e *Error3) Code() string {
    if e.flag {
        e.code = "examples-error-flagged"
    }
    if e.code == "" {
        return "examples-error-unknown"
    }
    return e.code
}
```

Whenever `Error3` is returned anywhere, the analysis adds the two constant error codes and the value assigned to the field as possible outcomes. A function containing the statement `return &Error3{false, "examples-error-not-implemented"}` has to declare the error codes: "examples-error-flagged", "examples-error-unknown", and "examples-error-not-implemented"

For more examples of possible or invalid error types see: [testdata/src/errortypes/](testdata/src/errortypes/)

## Error Code Origins

There are 3 possible origins of error codes that are considered:

1. [Type Construction](#type-construction)
2. [Assignment to Error Code Field](#assignment-to-error-code-field)
3. [Function Call](#function-call)

### Type Construction

```go
&Error{"examples-error-not-implemented"}
```

An error type construction adds all error codes that the `Code` method of this type may return. This includes:

* constant strings returned by `Code`
* error code field as assigned by the construction
  * In the example above the error code field is initialised to "examples-error-not-implemented" and this code is therefore added.
* constant strings assigned to the error code field in any method of the error type
  * See `Error3` above for an example, with the assignment: `e.code = "examples-error-flagged"`

### Assignment to Error Code Field

```go
err.code = "examples-error-closed"
```

Assigning a constant string to an error code field adds this string as an error code. If `err` in the example above is returned from a function, this function has to declare the error code "examples-error-closed".

For error codes assigned that way, the same format rules apply as for any other error code.

### Function Call

```go
TryOpen("example.txt")
```

```go
// Errors:
//
//    - examples-error-failed       -- failed to open file
//    - examples-error-invalid-name -- invalid file name
func TryOpen(fileName string) error { ... }
```

Calling a function adds all error codes which this function may return. In the example above, if the result of the function call to `TryOpen` is returned from a function, this function has to declare the error codes "examples-error-failed" and "examples-error-invalid-name".

Functions called from the **same package** as the caller are included in the analysis. This allows for **local functions** to not declare error codes but still work correctly in the analysis.

Calls to functions of **other packages** entierly trust the declared error codes. No messages are generated on the caller side, if declared and actual error codes have mismatches.

**Recursive calls** of functions set the error codes of all involved functions to the super set of error codes in those functions. See [testdata/src/recursion/recursion.go](testdata/src/recursion/recursion.go) for some examples.

## Interfaces

## Error Constructors

## Limitations

TODO:

* Dead branches cannot be detected
* Error has to be last return
* Leaking modifiable errors
