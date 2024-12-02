package sequencer

import "fmt"

type SequenceMethod string

// SequenceMethod constants
const (
	TimeStamp       SequenceMethod = "timestamp"
	OrderedChildren SequenceMethod = "ordered children"
	Combination     SequenceMethod = "combination"
)

// GetSequenceMethod returns the SequenceMethod for the given string.
// Returns an error if the string is not a valid SequenceMethod.
//
// Args:
// 1. s: string. The string to convert to a SequenceMethod.
//
// Returns:
// 1. SequenceMethod. The SequenceMethod for the given string.
// 2. error. An error if the string is not a valid SequenceMethod.
func GetSequenceMethod(s string) (SequenceMethod, error) {
	switch s {
	case "timestamp":
		return TimeStamp, nil
	case "ordered children":
		return OrderedChildren, nil
	case "combination":
		return Combination, nil
	default:
		return SequenceMethod(""), fmt.Errorf("unknown sequence method: %s", s)
	}
}
