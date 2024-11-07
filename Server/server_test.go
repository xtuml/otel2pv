package Server

import (
	"errors"
	"testing"
)

// MockServer is a mock implementation of the Server interface
type MockServer struct {
	isError bool
}

func (s *MockServer) Serve() error {
	if s.isError {
		return errors.New("test error")
	}
	return nil
}
func TestServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := &MockServer{
			isError: false,
		}

		err := server.Serve()

		if err != nil {
			t.Errorf("Expected no error from Serve, got %v", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		server := &MockServer{
			isError: true,
		}

		err := server.Serve()

		if err == nil {
			t.Errorf("Expected error from Serve, got '%v'", err)
		}
	})
}
