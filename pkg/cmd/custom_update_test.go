// Custom CLI extension code. Not generated.
package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompareReleaseVersions(t *testing.T) {
	tests := []struct {
		name     string
		left     string
		right    string
		expected int
	}{
		{name: "newer", left: "0.39.0", right: "0.38.0", expected: 1},
		{name: "older", left: "v0.37.9", right: "0.38.0", expected: -1},
		{name: "equal", left: "0.38.0", right: "v0.38.0", expected: 0},
		{name: "major takes precedence", left: "1.0.0", right: "0.99.99", expected: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := compareReleaseVersions(test.left, test.right)
			require.NoError(t, err)
			require.Equal(t, test.expected, got)
		})
	}
}

func TestParseChecksums(t *testing.T) {
	checksums := parseChecksums([]byte("aabbcc\n" +
		"d82fac78732b6106b332e527f7d962fb2e66ebda2d1aac8628aeb634b3da821a  boltz-api.tar.gz\n" +
		"not-a-checksum  ignored\n"))

	require.Equal(t, map[string]string{
		"boltz-api.tar.gz": "d82fac78732b6106b332e527f7d962fb2e66ebda2d1aac8628aeb634b3da821a",
	}, checksums)
}

func TestDownloadInstallAssetVerifiesChecksum(t *testing.T) {
	payload := []byte("release archive")
	hash := sha256.Sum256(payload)
	expectedChecksum := hex.EncodeToString(hash[:])
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	t.Cleanup(server.Close)

	asset := installAsset{Name: "archive.tar.gz", URL: server.URL}
	path, err := downloadInstallAsset(t.Context(), server.Client(), asset, expectedChecksum)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(path) })
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, payload, got)

	_, err = downloadInstallAsset(t.Context(), server.Client(), asset, strings.Repeat("0", sha256.Size*2))
	require.ErrorContains(t, err, "checksum mismatch")
}

func TestReplaceExecutableFromArchive(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "boltz-api")
	require.NoError(t, os.WriteFile(target, []byte("old binary"), 0o755))

	expectedName := "boltz-api"
	assetName := "boltz-api_0.39.0_macos_arm64.zip"
	archivePath := filepath.Join(tmpDir, assetName)
	if runtime.GOOS == "linux" {
		assetName = "boltz-api_0.39.0_linux_amd64.tar.gz"
		archivePath = filepath.Join(tmpDir, assetName)
	}
	if runtime.GOOS == "windows" {
		expectedName = "boltz-api.exe"
		assetName = "boltz-api_0.39.0_windows_amd64.zip"
		archivePath = filepath.Join(tmpDir, assetName)
	}

	require.NoError(t, writeTestUpdateArchive(archivePath, expectedName, []byte("new binary")))
	require.NoError(t, replaceExecutableFromArchive(archivePath, assetName, target))
	updated, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, []byte("new binary"), updated)
}

func writeTestUpdateArchive(path, binaryName string, content []byte) error {
	if filepath.Ext(path) == ".gz" {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		gzipWriter := gzip.NewWriter(file)
		tarWriter := tar.NewWriter(gzipWriter)
		if err := tarWriter.WriteHeader(&tar.Header{Name: "nested/" + binaryName, Mode: 0o755, Size: int64(len(content))}); err != nil {
			_ = file.Close()
			return err
		}
		if _, err := tarWriter.Write(content); err != nil {
			_ = file.Close()
			return err
		}
		if err := tarWriter.Close(); err != nil {
			_ = file.Close()
			return err
		}
		if err := gzipWriter.Close(); err != nil {
			_ = file.Close()
			return err
		}
		return file.Close()
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	zipWriter := zip.NewWriter(file)
	entry, err := zipWriter.Create("nested/" + binaryName)
	if err == nil {
		_, err = io.Copy(entry, bytes.NewReader(content))
	}
	if closeErr := zipWriter.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}
