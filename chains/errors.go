package chains

import "errors"

var (
	// ErrInvalidInputValues is returned when some expected input values keys to
	// a chain is missing.
	ErrInvalidInputValues = errors.New("missing keys in input values")
	// ErrInvalidOutputValues is returned when expected output keys to a chain does
	// not match the actual keys in the return output values map.
	ErrInvalidOutputValues = errors.New("missing keys in output values")
	// ErrMultipleInputsInRun is returned in the run function if the chain expects
	// more then one input values.
	ErrMultipleInputsInRun = errors.New("run not supported in chain with more then one expected input")
	// ErrMultipleOutputsInRun is returned in the run function if the chain expects
	// more then one output values.
	ErrMultipleOutputsInRun = errors.New("run not supported in chain with more then one expected output")
	// ErrMultipleOutputsInRun is returned in the run function if the chain returns
	// a value that is not a string.
	ErrWrongOutputTypeInRun = errors.New("run not supported in chain that returns value that is not string")
)
