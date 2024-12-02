package sequencer

import (
	"testing"
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