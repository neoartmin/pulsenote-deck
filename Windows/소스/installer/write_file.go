package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

func writeFile(path string, data []byte) error {
	if old, e := os.ReadFile(path); e == nil && sha256.Sum256(old) == sha256.Sum256(data) {
		return nil
	}
	// Write a complete file in the destination directory before replacing
	// the installed binary. Unique temporary names avoid stale update files.
	f, err := os.CreateTemp(filepath.Dir(path), ".pnd-install-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return fmt.Errorf("설치 파일 교체 실패 (%s): %w", path, err)
	}
	return nil
}
