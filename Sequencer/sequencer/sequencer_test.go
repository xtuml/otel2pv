package sequencer

import (
	"reflect"
	"testing"
	
	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

// Test SequenceMethod
func TestSequenceMethod(t *testing.T) {
	tests := []struct {
		name string
		want SequenceMethod
	}{
		{
			name: "timestamp",
			want: TimeStamp,
		},
		{
			name: "ordered children",
			want: OrderedChildren,
		},
		{
			name: "combination",
			want: Combination,
		},
	}
	incorrect := SequenceMethod("incorrect")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SequenceMethod(tt.name); got != tt.want {
				t.Errorf("SequenceMethod() = %v, want %v", got, tt.want)
			}
			// Test incorrect value
			if incorrect == tt.want {
				t.Errorf("SequenceMethod() = %v, want %v", incorrect, tt.want)
			}
		})
	}
}

// Test GetSequenceMethod
func TestGetSequenceMethod(t *testing.T) {
	tests := []struct {
		name    string
		want    SequenceMethod
		wantErr bool
	}{
		{
			name:    "timestamp",
			want:    TimeStamp,
			wantErr: false,
		},
		{
			name:    "ordered children",
			want:    OrderedChildren,
			wantErr: false,
		},
		{
			name:    "combination",
			want:    Combination,
			wantErr: false,
		},
		{
			name:    "incorrect",
			want:    SequenceMethod(""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSequenceMethod(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSequenceMethod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSequenceMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Tests SequencerConfig
func TestSequencerConfig(t *testing.T) {
	t.Run("ImplementsConfig", func(t *testing.T) {
		sc := &SequencerConfig{}
		if _, ok := interface{}(sc).(Server.Config); !ok {
			t.Errorf("SequencerConfig does not implement Config")
		}
	})
	t.Run("IngestConfig", func(t *testing.T) {
		tests := []struct {
			name    string
			config  map[string]any
			wantErr bool
		}{
			{
				name: "valid",
				config: map[string]any{
					"sequenceMethod": "timestamp",
					"inputSchema":    "input",
					"outputSchema":   "output",
					"inputToOutput":  map[string]any{"input": "output"},
				},
				wantErr: false,
			},
			{
				name: "invalid sequenceMethod",
				config: map[string]any{
					"sequenceMethod": 1,
					"inputSchema":    "input",
					"outputSchema":   "output",
					"inputToOutput":  map[string]any{"input": "output"},
				},
				wantErr: true,
			},
			{
				name: "invalid inputSchema",
				config: map[string]any{
					"sequenceMethod": "timestamp",
					"inputSchema":    1,
					"outputSchema":   "output",
					"inputToOutput":  map[string]any{"input": "output"},
				},
				wantErr: true,
			},
			{
				name: "invalid outputSchema",
				config: map[string]any{
					"sequenceMethod": "timestamp",
					"inputSchema":    "input",
					"outputSchema":   1,
					"inputToOutput":  map[string]any{"input": "output"},
				},
				wantErr: true,
			},
			{
				name: "invalid inputToOutput",
				config: map[string]any{
					"sequenceMethod": "timestamp",
					"inputSchema":    "input",
					"outputSchema":   "output",
					"inputToOutput":  map[string]any{"input": 1},
				},
				wantErr: true,
			},
			{
				name: "missing sequenceMethod",
				config: map[string]any{
					"inputSchema":    "input",
					"outputSchema":   "output",
					"inputToOutput":  map[string]any{"input": "output"},
				},
				wantErr: true,
			},
			{
				name: "missing inputSchema",
				config: map[string]any{
					"sequenceMethod": "timestamp",
					"outputSchema":   "output",
					"inputToOutput":  map[string]any{"input": "output"},
				},
				wantErr: true,
			},
			{
				name: "missing outputSchema",
				config: map[string]any{
					"sequenceMethod": "timestamp",
					"inputSchema":    "input",
					"inputToOutput":  map[string]any{"input": "output"},
				},
				wantErr: true,
			},
			{
				name: "missing inputToOutput",
				config: map[string]any{
					"sequenceMethod": "timestamp",
					"inputSchema":    "input",
					"outputSchema":   "output",
				},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				sc := &SequencerConfig{}
				if err := sc.IngestConfig(tt.config); (err != nil) != tt.wantErr {
					t.Errorf("SequencerConfig.IngestConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.name == "valid" {
					if sc.sequenceMethod != TimeStamp {
						t.Errorf("SequencerConfig.IngestConfig() sequenceMethod = %v, want %v", sc.sequenceMethod, TimeStamp)
					}
					if sc.inputSchema != "input" {
						t.Errorf("SequencerConfig.IngestConfig() inputSchema = %v, want %v", sc.inputSchema, "input")
					}
					if sc.outputSchema != "output" {
						t.Errorf("SequencerConfig.IngestConfig() outputSchema = %v, want %v", sc.outputSchema, "output")
					}
					if !reflect.DeepEqual(sc.inputToOutput, map[string]string{"input": "output"}) {
						t.Errorf("SequencerConfig.IngestConfig() inputToOutput = %v, want %v", sc.inputToOutput, map[string]string{"input": "output"})
					}
				}
			})
		}
	})
}