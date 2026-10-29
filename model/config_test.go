package model

import "testing"

func TestSetGetString(t *testing.T) {
	if got, err := GetConfig("key", "default"); err != nil || got != "default" {
		t.Fatalf("default: %q %v", got, err)
	}
	if err := SetConfig("key", "value"); err != nil {
		t.Fatal(err)
	}
	if got, err := GetConfig("key", "default"); err != nil || got != "value" {
		t.Fatalf("saved: %q %v", got, err)
	}
	if err := DelConfig("key"); err != nil {
		t.Fatal(err)
	}
	if got, err := GetConfig("key", "default"); err != nil || got != "default" {
		t.Fatalf("deleted: %q %v", got, err)
	}
}
