package main

import (
	"fmt"
	"os"
)

// Missing files are fine (repeated uninstall); access denied and locked
// files are NOT success and must be reported to the user.
func removeInstalledFile(path string) error {
	if info, err := os.Lstat(path); err == nil && info.IsDir() {
		return fmt.Errorf("파일 제거 실패 (%s): 예상하지 못한 디렉터리", path)
	}
	err := os.Remove(path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("파일 제거 실패 (%s): %w", path, err)
}
