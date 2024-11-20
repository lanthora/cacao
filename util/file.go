package util

import (
	"os"
	"path"
)

func FindFileByExtFromDir(dir string, ext string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if path.Ext(file.Name()) == ext {
			return file.Name(), nil
		}
	}
	return "", os.ErrNotExist
}
