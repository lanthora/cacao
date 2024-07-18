package argp

import (
	"os"
	"testing"
)

func TestGetString(t *testing.T) {
	want := "default"
	if got := Get("key", "default"); got != want {
		t.Fatalf(`GetString("key", "default") = %v, want %v`, got, want)
	}
	os.Args = append(os.Args, "--key=value")
	want = "value"
	if got := Get("key", "default"); got != want {
		t.Fatalf(`GetString("key", "default") = %v, want %v`, got, want)
	}
}
