package assert

import (
	"fmt"
)

// PanicErr if the error is not nil, panic the error
func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Assert assert that the checkFunc should be passed, or the program will panic error.
func Assert(argName string, assertFunc func() bool) {
	if !assertFunc() {
		panic(fmt.Errorf("Invalid value of parameter/argument '%s'", argName))
	}
}

// NotNil assert the data should not be nil
func NotNil(argName string, data interface{}) {
	Assert(argName, func() bool {
		return data != nil
	})
}

// Assert assert the string length should not be empty
func NotEmpty(argName, value string) {
	Assert(argName, func() bool {
		return len(value) > 0
	})
}
