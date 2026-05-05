package main

import "testing"

func TestParsePortUsesDefaultWhenNoArgs(t *testing.T) {
	port, err := parsePort(nil)
	if err != nil {
		t.Fatalf("parsePort returned error: %v", err)
	}
	if port != defaultPort {
		t.Fatalf("expected default port %q, got %q", defaultPort, port)
	}
}

func TestParsePortUsesProvidedPort(t *testing.T) {
	port, err := parsePort([]string{"2525"})
	if err != nil {
		t.Fatalf("parsePort returned error: %v", err)
	}
	if port != "2525" {
		t.Fatalf("expected port %q, got %q", "2525", port)
	}
}

func TestParsePortRejectsExtraArgs(t *testing.T) {
	_, err := parsePort([]string{"2525", "localhost"})
	if err == nil {
		t.Fatal("expected error for extra arguments")
	}
}
