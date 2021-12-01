# Error Code Analysis

## Install and Run Analyser

The analyser in this package checks:

* if error returning functions declare possible error codes.
* if the declared error codes match what is actually returned.

How to run the analyser as stand-alone tool:

* Installing go-serum-analyzer: `go install ./cmd/go-serum-analyzer`
* Change Directory to Target Project: `cd <target-path>`
* Execute Analyser: `go-serum-analyzer <package>`

## Command Line Options

### -strict

When set: requires functions to declare error codes. Only functions that return an error and that are exported are required to declare error codes. (See [Declaring and Returning Errors](#declaring-and-returning-errors))

When using the analyser in an IDE, we recommend that the **-strict** flag is generally turned on.

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

`*Error` implements the `error` interface and also is a Serum error because `interface { Code() string }` is implemented.
Those two conditions are the only ones needed to be considered a Serum error by the error code analysis.

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
  * `<code>` has to be composed of several hunks of ASCII text, which are joined by dash characters ("-"). More precisely `<code>` has to match `^[a-zA-Z][a-zA-Z0-9\-]*[a-zA-Z0-9]$` or `^[a-zA-Z]$`. Find more information on recommended conventions for codes under: [../README.md](../README.md#recommended-conventions-for-codes)
  * `<comment>` can be any string.
  * Any number of **whitespaces** are allowed before and after every part of a declaration.
    * In the example above whitespaces are used to align the double dashes ("--") after the `<code>` part.
    * It is highly recommended to use enough whitespace before the declaration to turn it into a preformatted block per golang's documentation conventions, as this will be the most readable when the docs are rendered to HTML.

* **Comments** are allowed between error code declarations:
  * Any line that does not start with a dash character ("-") is regarded as a comment.

* The **declaration block ends** when there's another fully blank line.

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
...\testdata\src\examples\02_basic_examples.go:38:1: function "AddUnused" has a mismatch of declared and actual error codes: unused codes: [examples-error-invalid-arg examples-error-invalid-collection examples-error-limit-reached]
```

The second example returns errors, but does not declare any of the returned codes.

```go
// Errors: none -- not actually true, but we want to showcase missing error codes.
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
...\testdata\src\examples\02_basic_examples.go:46:1: function "AddMissing" has a mismatch of declared and actual error codes: missing codes: [examples-error-invalid-arg examples-error-invalid-collection examples-error-limit-reached]
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

To be considered a valid Serum error, a type must implement the following interfaces:

* `error`, which is `interface { Error() string }`
* `interface { Code() string }`

For the `Error` method, there are no limitations to how it may be implemented. We recommend to follow the style laid out in: [../README.md](../README.md).

The `Code` method is used to get the error code from an error type instance. For the analysis tool to be able to statically track error codes, the `Code` method has to follow some conventions:

* return a constant string
  * Multiple different constant strings may be returned.
  * The returned strings have to be error codes of a valid format. (i.e. has to match `^[a-zA-Z][a-zA-Z0-9\-]*[a-zA-Z0-9]$` or `^[a-zA-Z]$`)
  * Empty strings are allowed, but not considered error codes.
* return a field
  * If a field is returned multiple times, always the same field has to be returned.
  * We call this field the **error code field**.
  * When constructing an error, this field has to be set to a constant string.
* return a local string variable
  * The only values that are allowed to be assigned to the returned variable are:
    * constant strings
    * fields
    * other local variables
  * Those values have to follow the same rules as described above for returned values.

There may be a mix of returning the error code field and constants inside of one `Code` method.

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

As a final example we show how returning fields and constants can be mixed:

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

## Annotations

Annotations can be used to overrule error code analysis.

The following annotations are only valid for **return statements** and should usually be placed directly above an statement returning an error with error code.

### Overwrite Codes Annotation

Overwrite annotation allow overwriting the error code analysis with own error codes.

```go
// Errors:
//
//    - examples-error-overwritten -- error forced by annotation
func Overwrite() error {
    // Error Codes = examples-error-overwritten
    return &Error{"examples-error-unknown"}
}

```

In the example above, the analyser believes, that the returned error code is "examples-error-overwritten".
Be aware, that the annotation does not change what is actually returned. In the example `Overwrite` would still return an error with code "examples-error-unknown".

The following example shows how an overwrite annotation could be used in practice:

```go
// Errors:
//
//    - examples-error-one -- error forced by annotation
//    - examples-error-two -- error forced by annotation
func OverwriteMultiple(index int) error { // want OverwriteMultiple:"ErrorCodes: examples-error-one examples-error-two"
    errors := []error{
        nil,
        &Error{"examples-error-one"},
        &Error{"examples-error-two"},
    }
    // Error Codes = examples-error-one, examples-error-two
    return errors[index]
}
```

The anylser cannot figure out that the codes "examples-error-one" and "examples-error-two" are returned by itself. Using the overwrite annotation the analyser is informed which error codes to expect.

Also note, that multiple comma-separated codes can be specified in the overwrite annotation.

### Add/Remove Codes Annotation

Add (resp. remove) annotations allow adding (removing) error codes to (from) the codes resulting from the analysis.

Adding error codes is done with the annotation `// Error Codes += code1, code2, ...` and removing analogously with `// Error Codes -= code1, code2, ...`.

```go
// Errors:
//
//    - examples-error-two   --
//    - examples-error-three --
func SubtractCode() error {
    switch err := MultipleCodes(); err.Code() {
    case "examples-error-one":
        HandleErrorOne(err)
        return nil
    default:
        // Error Codes -= examples-error-one
        return err
    }
}
```

The example above shows, how a remove annotation could be used to tell the analyser to ignore an error code that has already been handled.

Additionally there's an annotation from that allows adding and removing codes in the same line:

```go
// Error Codes -examples-error-two -examples-error-three +examples-error-extra
return err
```

In front of each error code is either a '+' or a '-', which indicates, if the error code should be added or removed.

**Important:** Remove code annotations are currently not supported for error codes with origin from error code field assignments. For example:

```go
err := &Error{}
err.TheCode = "assigned-error"
// Error Codes -= assigned-error
return err
```

In the example above the analyser would still consider "assigned-error" to be returned.

## Interfaces

Error codes can be declared for interface methods.

```go
type Box interface {
    // Put makes the box store the given value.
    //
    // Errors:
    //
    //    - examples-error-arg-nil -- if the given value is nil
    //    - examples-error-invalid -- if the box already holds a value
    //    - examples-error-unknown -- if an unexpected error occurred
    Put(value interface{}) error

    // Pop retrieves the value stored in the box and removes it from the box.
    //
    // Errors:
    //
    //    - examples-error-invalid -- if the box was empty
    Pop() (interface{}, error)
}
```

* All interface methods are required to declare error codes, if they return an error.
* The format of the declaration and the format of the error code is checked in the same way as for function definitions. (See more under [Error Declaration](#error-declaration))
* Calls to interface methods add the declared error codes to the analysis.

The following code snippets contain definitions for `BoxImpl` and `BoxInvalidImpl`, two types, which implement the `Box` interface:

```go
type BoxImpl struct {
    value interface{}
}

// Errors:
//
//    - examples-error-arg-nil -- if the given value is nil
//    - examples-error-invalid -- if the box already holds a value
func (b *BoxImpl) Put(value interface{}) error {
    if value == nil {
        return &Error{"examples-error-arg-nil"}
    }

    if b == nil || b.value != nil {
        return &Error{"examples-error-invalid"}
    }

    b.value = value
    return nil
}

// Errors:
//
//    - examples-error-invalid -- if the box was empty
func (b *BoxImpl) Pop() (interface{}, error) {
    if b == nil || b.value == nil {
        return nil, &Error{"examples-error-invalid"}
    }

    b.value = nil
    return b.value, nil
}
```

```go
type BoxInvalidImpl struct{}

// Errors:
//
//    - examples-error-not-implemented --
func (b *BoxInvalidImpl) Put(value interface{}) error {
    return &Error{"examples-error-not-implemented"}
}

// Errors:
//
//    - examples-error-not-implemented --
func (b *BoxInvalidImpl) Pop() (interface{}, error) {
    return nil, &Error{"examples-error-not-implemented"}
}
```

`BoxImpl` is a valid implementation of the `Box` interface.

Implementations are allowed to use fewer error codes than are present in the interface declaration. As shown with the error code "examples-error-unknown", which is not returned by the `Put` method of `BoxImpl`.

### Invalid Interface Implementation

Implementations are not allowed to declare error codes which are not present in the interface declaration. `BoxInvalidImpl` violates this requirement for both methods.

But the analyser does not complain at all about `BoxInvalidImpl`. This is the case, because the analyser doesn't actually know, if `BoxInvalidImpl` was supposed to implement `Box` or not. The analyser handles this case similar to the go compiler and only emits errors if `BoxInvalidImpl` is tried to be used as `Box`.

For example the following scenario:

```go
func UseBoxInvalidImplAsBox() {
    var b Box = &BoxInvalidImpl{}
    b.Put(b)
}
```

Here `BoxInvalidImpl` is assigned to a variable of type `Box` and therefore the analyser outputs the following lines:

```text
...\testdata\src\examples\04_interfaces.go:80:14: cannot use expression as "Box" value: method "Pop" declares the following error codes which were not part of the interface: [examples-error-not-implemented]
...\testdata\src\examples\04_interfaces.go:80:14: cannot use expression as "Box" value: method "Put" declares the following error codes which were not part of the interface: [examples-error-not-implemented]
```

### Embedding Interfaces

If an embedding interface or any of it's embedded interfaces contain error returning methods, the analyser checks if the declared error codes for those methods are compatible. This means concretely: if two methods with the same name are found, their declared error codes have to exactly match.

This approach is insipred by how the go compiler handles embedding interfaces: there, if two methods with the same name are found, their signature has to exactly match.

The following example demonstrates, how embedding interfaces are handled in the analysis.

```go
type Box interface {
    // Errors:
    //
    //    - examples-error-arg-nil -- if the given value is nil
    //    - examples-error-invalid -- if the box already holds a value
    //    - examples-error-unknown -- if an unexpected error occurred
    Put(value interface{}) error

    // Errors:
    //
    //    - examples-error-invalid -- if the box was empty
    Pop() (interface{}, error)
}

type Box2 interface {
    // Errors: none
    Put(value interface{}) error

    // Errors:
    //
    //    - examples-error-invalid -- in case of an invalid operation
    Pop() (interface{}, error)
}

type EmbeddingBox interface {
    Box
    Box2
}
```

The interface `EmbeddingBox` embedds both `Box` and `Box2`. The error codes of the methods `Box.Pop` and `Box2.Pop` match exactly and are therefore no problem. The declared error codes of the `Put` methods do not match and the analyser produces the following error message:

```text
...\testdata\src\examples\04_interfaces.go:103:2: embedded interface is not compatible: method "Put" has mismatches in declared error codes: missing codes: [examples-error-arg-nil examples-error-invalid examples-error-unknown]
```

## Error Constructors

The analysis tool allows the definition of error constructors:

* Error constructors declare exactly one parameter to be an error code.
  * We call this parameter: **error code parameter**
  * The declaration has to be inside the error declarations block,
  * and match the format `- param: <param-name> -- <comment>`.
    * `<param-name>` has to be a function parameter.
    * That parameter has to be of type `string`.
* The error code parameter can then be used wherever a constant string error code is used.
* When calling an error constructor, the error code argument has to be a constant string or an error code parameter.

The following examples illustrate how to define error constructors and how to use them:

```go
// Errors:
//
//    - param: code -- error code parameter
func NewError(code string) *Error {
    return &Error{code}
}

// Errors:
//
//    - examples-error-not-implemented --
func CallConstructor() error {
    return NewError("examples-error-not-implemented")
}
```

Assignments of error code parameters to error code fields are handled by the analyser.

```go
// Errors:
//
//    - param: c               --
//    - examples-error-unknown --
//    - examples-error-flagged --
func NewError3(flag bool, c string) *Error3 {
    err := &Error3{flag, "examples-error-unknown"}
    if flag {
        err.code = c
    }
    return err
}
```

Error constructors may also call other error constructors and even recursive calls of error constructors are allowed.

```go
// Errors:
//
//    - param: errorCode       --
//    - examples-error-unknown --
//    - examples-error-flagged --
func NewGeneralError(flag bool, errorCode string) error {
    if flag {
        return NewError3(flag, errorCode)
    }
    return NewError(errorCode)
}
```

!!!The following check is not yet implemented!!!

Error constructors are not allowed to modify the error code parameter, pass it to functions, or use it in type construction. This limitation is enforced, to make static analysis possible. (E.g. a function could modify the error code parameter without us knowing, and we want to avoid that.)

## Limitations

This section describes limitations in the analyser. That includes:

* things that cannot be checked, but should be avoided anyways
* things that are restricted, to make the design and implementation of the analyser simple

### Leaking Modifiable Errors

It is allowed to leak modifiable errors.

For example in the following code, the invalid assignment to the error code would not be found by the analysis tool:

```go
// Errors:
//
//    - examples-error-invalid --
func CallModifyError() error {
    err := &Error{"examples-error-invalid"}
    ModifyError(err)
    return err
}

func ModifyError(err *Error) {
    err.TheCode = "some invalid value"
}
```

This is mostly the case to allow clients to pass errors to logging functions and similar. For example, the following could be a desirable scenario:

```go
var err *Error = TryOpen(fileName)
if err != nil {
    logger.Fatalf(err, "failed to open file: %q", fileName)
}
```

### Dead Branches Not Detected

The analysis does not consider any branches. The error code analysis calculates the super set of possible error codes in a function. This is done by visiting every branch and collecting all error codes everywhere.

The following example demonstrates this limit:

```go
// Errors:
//
//    - example-error-unreachable -- is never actually returned
func DeadBranchError() error {
    if false {
        return &Error{"example-error-unreachable"}
    }
    return nil
}
```

### Error has to be Last Result

When a function has multiple results, the error result has to be the last result. This is a convention that is already common (but not enforced) in go and simplifies the analysis.

The tool recognises if a function has the error as a non-last return and complains about it.

```go
func ErrorNotLast() (error, string) {
    return nil, nil
}
```

The example code above would result in the following error message from the analyser:

```text
...\testdata\src\examples\06_limitations.go:26:22: error should be returned as the last argument
```
