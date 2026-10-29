package model

import (
	"fmt"
	"os"
	"testing"

	"github.com/lanthora/cacao/storage"
)

func TestMain(m *testing.M) {
	os.Exit(func() int {
		directory, err := os.MkdirTemp("", "cacao-model-tests-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer os.RemoveAll(directory)
		if err := storage.Initialize(directory); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		pool, err := storage.Get().DB()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer pool.Close()
		if err := Initialize(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return m.Run()
	}())
}
