# Error Code Analysis

The analyser in this package checks:

* if error returning functions declare possible error codes. (This only holds for exported function.)
* if the declared error codes match what is actually returned.

Run the analyser as stand-alone tool by:

* Installing reeverify: `go install cmd/analyse/reeverify.go`
* Change Directory to Target Project: `cd <target-path>`
* Execute Analyser: `reeverify <package>`

## Examples

The following code shows an example of an exported function Exported() with declared error codes. [...]
