package fileorigin

import (
	"io"
	"os"
	"path/filepath"
)

func copyFile(srcpath string, dstpath string) error {
	parent := filepath.Dir(dstpath)
	r, err := os.Open(srcpath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(parent, os.ModePerm)
	if err != nil {
		return err
	}
	w, err := os.Create(dstpath)
	if err != nil {
		return err
	}

	defer func() {
		if c := w.Close(); err == nil {
			err = c
		}
	}()

	_, err = io.Copy(w, r)
	return err
}
