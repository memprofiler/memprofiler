package test

import (
	"testing"
)

func TestIntegration(t *testing.T) {

	// run memprofiler server
	sl := newServerLauncher()
	if err := sl.start(); err != nil {
		t.Fatal("Failed to start server", err)
	}
	defer func() {
		if err := sl.stop(); err != nil {
			t.Fatal("Failed to stop server", err)
		}
	}()
}
