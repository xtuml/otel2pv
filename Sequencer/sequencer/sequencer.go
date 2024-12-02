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

// SequencerConfig is a struct that represents the configuration for a Sequencer.
// It has the following fields:
//
// 1. sequenceMethod: SequenceMethod. The method to use for sequencing.
//
// 2. inputSchema: string. The schema for the input data.
//
// 3. outputSchema: string. The schema for the output data.
//
// 4. inputToOutput: map[string]string. A map of input schema fields to output schema fields.
type SequencerConfig struct {
	sequenceMethod SequenceMethod
	inputSchema    string
	outputSchema   string
	inputToOutput  map[string]string
}

// IngestConfig ingests the configuration for the SequencerConfig.
// Returns an error if the configuration is invalid.
//
// Args:
// 1. config: map[string]any. The raw configuration for the Sequencer.
//
// Returns:
// 1. error. An error if the configuration is invalid.
func (sc *SequencerConfig) IngestConfig(config map[string]any) error {
	sequenceMethod, ok := config["sequenceMethod"].(string)
	if !ok {
		return fmt.Errorf("invalid sequenceMethod - must be a string and must be set")
	}
	sm, err := GetSequenceMethod(sequenceMethod)
	if err != nil {
		return fmt.Errorf("invalid sequenceMethod - %s", err)
	}
	sc.sequenceMethod = sm

	inputSchema, ok := config["inputSchema"].(string)
	if !ok {
		return fmt.Errorf("invalid inputSchema - must be a string and must be set")
	}
	sc.inputSchema = inputSchema

	outputSchema, ok := config["outputSchema"].(string)
	if !ok {
		return fmt.Errorf("invalid outputSchema - must be a string and must be set")
	}
	sc.outputSchema = outputSchema

	inputToOutput, ok := config["inputToOutput"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid inputToOutput - must be a map of field to field")
	}
	sc.inputToOutput = make(map[string]string)
	for key, value := range inputToOutput {
		outputField, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid inputToOutput - must map a string identifier to a string")
		}
		sc.inputToOutput[key] = outputField
	}
	return nil
}