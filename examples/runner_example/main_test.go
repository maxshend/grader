package main

import "testing"

func TestHello(t *testing.T) {
	expect := "Hello, World!"
	got := Hello()

	if expect != got {
		t.Errorf("expected %q, got %q", expect, got)
	}
}
