// Custom CLI extension code. Not generated.
//go:build windows

package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf16"
)

func atomicReplaceFile(source, target string, mode os.FileMode) error {
	pending, err := os.CreateTemp(filepath.Dir(target), ".boltz-api-update-pending-*")
	if err != nil {
		return fmt.Errorf("create pending replacement beside %s: %w", target, err)
	}
	pendingPath := pending.Name()
	defer func() {
		_ = pending.Close()
		if err != nil {
			_ = os.Remove(pendingPath)
		}
	}()

	input, err := os.Open(source)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(pending, input)
	_ = input.Close()
	if copyErr != nil {
		return copyErr
	}
	if err = pending.Chmod(mode); err != nil {
		return err
	}
	if err = pending.Close(); err != nil {
		return err
	}

	command := fmt.Sprintf(
		"$p = Get-Process -Id %d -ErrorAction SilentlyContinue; if ($p) { Wait-Process -Id %d }; Move-Item -LiteralPath '%s' -Destination '%s' -Force",
		os.Getpid(),
		os.Getpid(),
		strings.ReplaceAll(pendingPath, "'", "''"),
		strings.ReplaceAll(target, "'", "''"),
	)
	encoded := base64.StdEncoding.EncodeToString(utf16Bytes(command))
	helper := exec.Command("powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-EncodedCommand", encoded)
	helper.Stdin = nil
	helper.Stdout = io.Discard
	helper.Stderr = io.Discard
	if err := helper.Start(); err != nil {
		return fmt.Errorf("start Windows update helper: %w", err)
	}
	return nil
}

func utf16Bytes(value string) []byte {
	encoded := utf16.Encode([]rune(value))
	bytes := make([]byte, 0, len(encoded)*2)
	for _, value := range encoded {
		bytes = append(bytes, byte(value), byte(value>>8))
	}
	return bytes
}
