package funcliteral

func namedFunctionInOtherFile() error {
	return &Error{"other-function-error"}
}
