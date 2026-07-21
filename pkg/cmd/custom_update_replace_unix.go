// Custom CLI extension code. Not generated.
//go:build !windows

package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func atomicReplaceFile(source, target string, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(target), ".boltz-api-update-*")
	if err != nil {
		return fmt.Errorf("create replacement beside %s: %w", target, err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	input, err := os.Open(source)
	if err != nil {
		_ = tmp.Close()
		return err
	}
	_, copyErr := io.Copy(tmp, input)
	_ = input.Close()
	if copyErr != nil {
		_ = tmp.Close()
		return copyErr
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, target); err != nil {
		return fmt.Errorf("replace %s: %w", target, err)
	}
	return nil
}
