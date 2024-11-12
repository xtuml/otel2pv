package jqextractor

import (
	"testing"

	"github.com/itchyny/gojq"

	"github.com/SmartDCSITlimited/CDS-OTel-To-PV/Server"
)

type MockReceiver struct {
	// MockReceiver is a struct that provides a mock implementation of the Receiver interface
	mychan chan Server.AppData
}

func (mr *MockReceiver) SendTo(data Server.AppData) error {
	mr.mychan <- data
	return nil
}

func (mr *MockReceiver) GetOutChan() (<-chan Server.AppData, error) {
	return mr.mychan, nil
}

func TestJQTransformer(t *testing.T) {
	jqQuery, err := gojq.Parse(".")
	if err != nil {
		t.Errorf("Error parsing JQ program: %v", err)
	}
	jqProgram, err := gojq.Compile(jqQuery)
	if err != nil {
		t.Errorf("Error compiling JQ program: %v", err)
	}
	inChan := make(chan Server.AppData)
	outChan := make(chan Server.AppData)
	inReceiver := MockReceiver{
		mychan: inChan,
	}
	outReceiver := MockReceiver{
		mychan: outChan,
	}
	t.Run("Instantiation", func(t *testing.T) {
		jqTransformer := JQTransformer{
			jqProgram:   jqProgram,
			inReceiver:  &inReceiver,
			outReceiver: &outReceiver,
		}
		if jqTransformer.jqProgram != jqProgram {
			t.Errorf("Expected jqProgram to be %v, got %v", jqProgram, jqTransformer.jqProgram)
		}
		if jqTransformer.inReceiver != &inReceiver {
			t.Errorf("Expected inReceiver to be %v, got %v", &inReceiver, jqTransformer.inReceiver)
		}
		if jqTransformer.outReceiver != &outReceiver {
			t.Errorf("Expected outReceiver to be %v, got %v", &outReceiver, jqTransformer.outReceiver)
		}
	})
	t.Run("GetReceiver", func(t *testing.T) {
		jqTransformer := JQTransformer{jqProgram: jqProgram}
		// Test when inReceiver is nil
		_, err := jqTransformer.GetReceiver()
		if err == nil {
			t.Errorf("Expected error from GetReceiver, got nil")
		}
		// Test when inReceiver is set
		jqTransformer.inReceiver = &inReceiver
		receiver, err := jqTransformer.GetReceiver()
		if err != nil {
			t.Errorf("Expected no error from GetReceiver, got %v", err)
		}
		if receiver != &inReceiver {
			t.Errorf("Expected receiver to be %v, got %v", &inReceiver, receiver)
		}
	})
	t.Run("AddReceiver", func(t *testing.T) {
		jqTransformer := JQTransformer{jqProgram: jqProgram}
		// Test when outReceiver is nil
		err := jqTransformer.AddReceiver(&outReceiver)
		if err != nil {
			t.Errorf("Expected no error from AddReceiver, got %v", err)
		}
		if jqTransformer.outReceiver != &outReceiver {
			t.Errorf("Expected outReceiver to be %v, got %v", &outReceiver, jqTransformer.outReceiver)
		}
		// Test when outReceiver is set
		err = jqTransformer.AddReceiver(&outReceiver)
		if err == nil {
			t.Errorf("Expected error from AddReceiver, got nil")
		}
	})
}
