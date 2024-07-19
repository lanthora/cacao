package model

import (
	"testing"
)

func TestSetGetString(t *testing.T) {
	if got, ok := GetString("key"); ok {
		t.Fatalf(`GetString("key"") = %v, %v, want %v %v`, got, ok, "", false)
	}
	SetString("key", "value")
	if got, ok := GetString("key"); !ok || got != "value" {
		t.Fatalf(`GetString("key"") = %v, %v, want %v %v`, got, ok, "value", true)
	}
}
