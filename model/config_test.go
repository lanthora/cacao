package model

import (
	"testing"
)

func TestSetGetString(t *testing.T) {
	if got := GetConfig("key", "default"); got != "default" {
		t.Fatalf(`GetString("key") = %v, want %v`, got, "default")
	}
	SetConfig("key", "value")
	if got := GetConfig("key", "default"); got != "value" {
		t.Fatalf(`GetString("key") = %v, want %v`, got, "value")
	}
	DelConfig("key")
	if got := GetConfig("key", "default"); got != "default" {
		t.Fatalf(`GetString("key") = %v, want %v`, got, "default")
	}
}
