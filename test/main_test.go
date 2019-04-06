package test

import (
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	l, err := newLauncher()
	if err != nil {
		t.Fatal(err)
	}
	defer l.Stop()

	time.Sleep(3 * time.Second)
}
