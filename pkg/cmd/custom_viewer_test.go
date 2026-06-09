// Custom CLI extension code. Not generated.
package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewerOpenServerManifestIncludesSupportedFiles(t *testing.T) {
	rootDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "design.cif"), []byte("data_design\n#"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "design.npz"), []byte("npz"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "notes.txt"), []byte("ignore"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(rootDir, "nested"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "nested", "binder.pdb"), []byte("MODEL\n"), 0644))

	server, err := startViewerOpenServer(viewerOpenSpec{
		RootDir:      rootDir,
		WorkspaceDir: filepath.Dir(rootDir),
		ViewerURL:    viewerOpenDefaultURL,
	})
	require.NoError(t, err)
	defer server.Close()

	client := http.Client{Timeout: time.Second}
	request, err := http.NewRequest(http.MethodGet, server.manifestURL, nil)
	require.NoError(t, err)
	request.Header.Set("Origin", "https://lab.boltz.bio")
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "https://lab.boltz.bio", response.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", response.Header.Get("Access-Control-Allow-Private-Network"))

	var manifest viewerOpenManifest
	require.NoError(t, json.NewDecoder(response.Body).Decode(&manifest))

	require.Equal(t, viewerOpenManifestVersion, manifest.SchemaVersion)
	assert.Equal(t, filepath.Base(rootDir), manifest.Name)
	assert.NotEmpty(t, manifest.Workspace.Key)
	require.Len(t, manifest.Files, 3)
	assert.ElementsMatch(t, []string{"design.cif", "design.npz", "nested/binder.pdb"}, viewerOpenFileManifestPaths(manifest.Files))
}

func TestViewerOpenServerServesOnlyManifestFilesWithCORS(t *testing.T) {
	rootDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "design.cif"), []byte("data_design\n#"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(rootDir, "notes.txt"), []byte("ignore"), 0644))

	server, err := startViewerOpenServer(viewerOpenSpec{
		RootDir:      rootDir,
		WorkspaceDir: filepath.Dir(rootDir),
		ViewerURL:    viewerOpenDefaultURL,
	})
	require.NoError(t, err)
	defer server.Close()

	require.Len(t, server.files, 1)
	client := http.Client{Timeout: time.Second}
	request, err := http.NewRequest(http.MethodGet, server.files[0].URL, nil)
	require.NoError(t, err)
	request.Header.Set("Origin", "https://lab.boltz.bio")
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "https://lab.boltz.bio", response.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "chemical/x-cif", response.Header.Get("Content-Type"))

	ignoredRequest, err := http.NewRequest(http.MethodGet, server.baseURL+"/t/"+server.token+"/files/notes.txt", nil)
	require.NoError(t, err)
	ignoredResponse, err := client.Do(ignoredRequest)
	require.NoError(t, err)
	defer ignoredResponse.Body.Close()

	assert.Equal(t, http.StatusNotFound, ignoredResponse.StatusCode)
}

func TestBuildViewerOpenURLCarriesManifestAndWorkspace(t *testing.T) {
	workspace := viewerOpenManifestWorkspace{Key: "workspace-key", Name: "outputs"}
	viewerURL, origin, err := buildViewerOpenURL(
		"http://localhost:5199/viewer?existing=1",
		"http://127.0.0.1:43210/t/token/manifest.json",
		workspace,
	)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:5199", origin)
	assert.Contains(t, viewerURL, "existing=1")
	assert.Contains(t, viewerURL, "manifestUrl=http%3A%2F%2F127.0.0.1%3A43210%2Ft%2Ftoken%2Fmanifest.json")
	assert.Contains(t, viewerURL, "workspaceKey=workspace-key")
	assert.Contains(t, viewerURL, "workspaceName=outputs")
}

func viewerOpenFileManifestPaths(files []viewerOpenManifestFile) []string {
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths
}
