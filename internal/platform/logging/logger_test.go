package logging_test

import (
	"log/slog"
	"oci-runtime/internal/platform/logging"
	"testing"
)

func TestNew_ReturnsLogger(t *testing.T) {
	log := logging.New("dev")

	if log == nil {
		t.Fatal("New returned nil")
	}

	h := log.Handler()
	if _, ok := h.(slog.Handler); !ok {
		t.Fatalf("expected JSONHandler, got %T", h)
	}
}
