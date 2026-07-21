// Custom CLI extension code. Not generated.
package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	defaultInstallBaseURL = "https://install.boltz.bio/boltz-api"
	installBaseURLEnv     = "BOLTZ_API_INSTALL_BASE_URL"
	updateHTTPTimeout     = 30 * time.Second
	maxUpdateArchiveSize  = 512 << 20
	maxUpdateBinarySize   = 256 << 20
)

type installRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []installAsset `json:"assets"`
}

type installAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

var updateCommand = cli.Command{
	Name:            "update",
	Usage:           "Update boltz-api to the latest published release",
	HideHelpCommand: true,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "check",
			Usage: "Check for an update without downloading or replacing the CLI",
		},
	},
	Action: handleUpdate,
}

func handleUpdate(ctx context.Context, cmd *cli.Command) error {
	baseURL, err := installBaseURL()
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: updateHTTPTimeout}
	release, checksums, err := fetchInstallRelease(ctx, client, baseURL)
	if err != nil {
		return err
	}

	currentVersion := Version
	latestVersion, err := releaseVersion(release.TagName)
	if err != nil {
		return fmt.Errorf("invalid latest boltz-api release %q: %w", release.TagName, err)
	}
	comparison, err := compareReleaseVersions(latestVersion, currentVersion)
	if err != nil {
		return err
	}

	output := commandOutput(cmd)
	if comparison <= 0 {
		_, _ = fmt.Fprintf(output, "boltz-api %s is already the latest installed release\n", currentVersion)
		return nil
	}
	if cmd.Bool("check") {
		_, _ = fmt.Fprintf(output, "boltz-api update available: %s (current %s)\n", latestVersion, currentVersion)
		return nil
	}

	target, err := currentExecutablePath()
	if err != nil {
		return err
	}
	assetName, err := currentPlatformAssetName(latestVersion)
	if err != nil {
		return err
	}
	asset, ok := findInstallAsset(release.Assets, assetName)
	if !ok {
		return fmt.Errorf("latest boltz-api release %s has no asset for %s", latestVersion, assetName)
	}
	if err := validateInstallAssetURL(baseURL, asset.URL); err != nil {
		return err
	}
	expectedChecksum, ok := checksums[asset.Name]
	if !ok {
		return fmt.Errorf("release metadata has no checksum for %s", asset.Name)
	}

	_, _ = fmt.Fprintf(output, "Updating boltz-api from %s to %s...\n", currentVersion, latestVersion)
	archivePath, err := downloadInstallAsset(ctx, client, asset, expectedChecksum)
	if err != nil {
		return err
	}
	defer os.Remove(archivePath)

	if err := replaceExecutableFromArchive(archivePath, asset.Name, target); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(output, "Updated boltz-api to %s at %s\n", latestVersion, target)
	return nil
}

func installBaseURL() (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv(installBaseURLEnv)), "/")
	if baseURL == "" {
		baseURL = defaultInstallBaseURL
	}

	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil {
		return "", fmt.Errorf("%s must be an http(s) URL without credentials", installBaseURLEnv)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%s must use http or https", installBaseURLEnv)
	}
	return baseURL, nil
}

func fetchInstallRelease(ctx context.Context, client *http.Client, baseURL string) (installRelease, map[string]string, error) {
	var release installRelease
	if err := fetchJSON(ctx, client, baseURL+"/latest.json", &release); err != nil {
		return installRelease{}, nil, fmt.Errorf("fetch latest boltz-api release: %w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return installRelease{}, nil, errors.New("latest boltz-api metadata did not include a release tag")
	}

	tag := strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	checksumURL := fmt.Sprintf("%s/releases/v%s/checksums.txt", baseURL, tag)
	body, err := fetchBytes(ctx, client, checksumURL, 1<<20)
	if err != nil {
		return installRelease{}, nil, fmt.Errorf("fetch checksums for boltz-api %s: %w", release.TagName, err)
	}
	return release, parseChecksums(body), nil
}

func fetchJSON(ctx context.Context, client *http.Client, endpoint string, target any) error {
	body, err := fetchBytes(ctx, client, endpoint, 1<<20)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode %s: %w", endpoint, err)
	}
	return nil
}

func fetchBytes(ctx context.Context, client *http.Client, endpoint string, maxSize int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/plain")
	req.Header.Set("User-Agent", fmt.Sprintf("Boltz/CLI %s", Version))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s returned HTTP %d", endpoint, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxSize {
		return nil, fmt.Errorf("response from %s exceeds %d bytes", endpoint, maxSize)
	}
	return body, nil
}

func parseChecksums(body []byte) map[string]string {
	checksums := make(map[string]string)
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || len(fields[0]) != sha256.Size*2 {
			continue
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			continue
		}
		checksums[fields[1]] = strings.ToLower(fields[0])
	}
	return checksums
}

func findInstallAsset(assets []installAsset, name string) (installAsset, bool) {
	for _, asset := range assets {
		if asset.Name == name && strings.TrimSpace(asset.URL) != "" {
			return asset, true
		}
	}
	return installAsset{}, false
}

func validateInstallAssetURL(baseURL, assetURL string) error {
	base, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	asset, err := url.Parse(assetURL)
	if err != nil || asset.Scheme != base.Scheme || asset.Host != base.Host || asset.User != nil {
		return fmt.Errorf("release asset URL must use the configured install host")
	}
	return nil
}

func releaseVersion(tag string) (string, error) {
	version := strings.TrimPrefix(strings.TrimSpace(tag), "v")
	parts := strings.SplitN(version, "-", 2)[0]
	components := strings.Split(parts, ".")
	if len(components) != 3 {
		return "", fmt.Errorf("release version %q is not MAJOR.MINOR.PATCH", tag)
	}
	for _, component := range components {
		if _, err := strconv.Atoi(component); err != nil {
			return "", fmt.Errorf("release version %q is not numeric", tag)
		}
	}
	return version, nil
}

func compareReleaseVersions(left, right string) (int, error) {
	leftVersion, err := releaseVersion(left)
	if err != nil {
		return 0, err
	}
	rightVersion, err := releaseVersion(right)
	if err != nil {
		return 0, fmt.Errorf("current CLI version %q is invalid: %w", right, err)
	}

	leftParts := strings.Split(leftVersion, ".")
	rightParts := strings.Split(rightVersion, ".")
	for i := range leftParts {
		leftNumber, _ := strconv.Atoi(leftParts[i])
		rightNumber, _ := strconv.Atoi(rightParts[i])
		if leftNumber > rightNumber {
			return 1, nil
		}
		if leftNumber < rightNumber {
			return -1, nil
		}
	}
	return 0, nil
}

func currentPlatformAssetName(version string) (string, error) {
	platform := runtime.GOOS
	arch := runtime.GOARCH
	extension := "zip"
	switch platform {
	case "darwin":
		platform = "macos"
	case "linux":
		platform = "linux"
		extension = "tar.gz"
	case "windows":
		platform = "windows"
	default:
		return "", fmt.Errorf("automatic updates are not supported on %s", runtime.GOOS)
	}
	if arch == "arm" {
		arch = "armv6"
	}
	if arch != "386" && arch != "amd64" && arch != "arm64" && arch != "armv6" {
		return "", fmt.Errorf("automatic updates are not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Sprintf("boltz-api_%s_%s_%s.%s", version, platform, arch, extension), nil
}

func downloadInstallAsset(ctx context.Context, client *http.Client, asset installAsset, expectedChecksum string) (string, error) {
	tmp, err := os.CreateTemp("", "boltz-api-update-archive-*")
	if err != nil {
		return "", err
	}
	path := tmp.Name()
	removeOnError := true
	defer func() {
		_ = tmp.Close()
		if removeOnError {
			_ = os.Remove(path)
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("Boltz/CLI %s", Version))
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("download %s returned HTTP %d", asset.Name, resp.StatusCode)
	}

	hash := sha256.New()
	limited := io.LimitReader(resp.Body, maxUpdateArchiveSize+1)
	if _, err := io.Copy(io.MultiWriter(tmp, hash), limited); err != nil {
		return "", err
	}
	if info, err := tmp.Stat(); err != nil {
		return "", err
	} else if info.Size() > maxUpdateArchiveSize {
		return "", fmt.Errorf("downloaded archive %s exceeds %d bytes", asset.Name, maxUpdateArchiveSize)
	}
	if got := hex.EncodeToString(hash.Sum(nil)); !strings.EqualFold(got, expectedChecksum) {
		return "", fmt.Errorf("checksum mismatch for %s", asset.Name)
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	removeOnError = false
	return path, nil
}

func currentExecutablePath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate current boltz-api executable: %w", err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return "", err
	}
	linkInfo, err := os.Lstat(executable)
	if err == nil && linkInfo.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("current boltz-api is installed through a symlink; update it with the owning package manager or installer")
	}
	normalizedPath := filepath.ToSlash(executable)
	for _, marker := range []string{"/Cellar/", "/nix/store/"} {
		if strings.Contains(normalizedPath, marker) {
			return "", fmt.Errorf("current boltz-api appears to be package-manager managed; update it with the owning package manager")
		}
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err == nil {
		executable = resolved
	}
	info, err := os.Stat(executable)
	if err != nil {
		return "", fmt.Errorf("inspect current boltz-api executable: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("current boltz-api executable is not a regular file: %s", executable)
	}
	return executable, nil
}

func replaceExecutableFromArchive(archivePath, assetName, target string) error {
	tmpDir, err := os.MkdirTemp("", "boltz-api-update-extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	binaryPath := filepath.Join(tmpDir, "boltz-api")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	if err := extractUpdateBinary(archivePath, assetName, binaryPath); err != nil {
		return err
	}

	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	mode := info.Mode().Perm()
	if mode&0o111 == 0 {
		mode = 0o755
	}
	return atomicReplaceFile(binaryPath, target, mode)
}

func extractUpdateBinary(archivePath, assetName, outputPath string) error {
	expectedName := "boltz-api"
	if runtime.GOOS == "windows" {
		expectedName += ".exe"
	}
	if strings.HasSuffix(assetName, ".tar.gz") {
		archive, err := os.Open(archivePath)
		if err != nil {
			return err
		}
		defer archive.Close()
		gzipReader, err := gzip.NewReader(archive)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		tarReader := tar.NewReader(gzipReader)
		for {
			header, err := tarReader.Next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return err
			}
			if header.Typeflag == tar.TypeReg && path.Base(header.Name) == expectedName {
				return writeExtractedBinary(outputPath, io.LimitReader(tarReader, maxUpdateBinarySize+1))
			}
		}
		return fmt.Errorf("archive did not contain %s", expectedName)
	}

	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.FileInfo().IsDir() || path.Base(file.Name) != expectedName {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return err
		}
		err = writeExtractedBinary(outputPath, io.LimitReader(reader, maxUpdateBinarySize+1))
		_ = reader.Close()
		return err
	}
	return fmt.Errorf("archive did not contain %s", expectedName)
}

func writeExtractedBinary(path string, source io.Reader) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o700)
	if err != nil {
		return err
	}
	defer file.Close()
	size, err := io.Copy(file, source)
	if err != nil {
		return err
	}
	if size > maxUpdateBinarySize {
		return fmt.Errorf("extracted boltz-api binary exceeds %d bytes", maxUpdateBinarySize)
	}
	return file.Chmod(0o700)
}

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
		if runtime.GOOS == "windows" {
			return fmt.Errorf("replace %s: %w; close boltz-api and rerun the installer", target, err)
		}
		return fmt.Errorf("replace %s: %w", target, err)
	}
	return nil
}

func commandOutput(cmd *cli.Command) io.Writer {
	if cmd != nil {
		root := cmd.Root()
		if root != nil && root.Writer != nil {
			return root.Writer
		}
		if cmd.Writer != nil {
			return cmd.Writer
		}
	}
	return os.Stdout
}
