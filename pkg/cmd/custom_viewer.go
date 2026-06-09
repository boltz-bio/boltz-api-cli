// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	viewerOpenDefaultURL       = "https://lab.boltz.bio/viewer"
	viewerOpenManifestVersion  = 1
	viewerOpenTokenBytes       = 18
	viewerOpenLocalhostAddress = "127.0.0.1"
)

var viewerOpenStructureExtensions = []string{".cif.gz", ".cif", ".bcif", ".pdb"}
var viewerOpenSidecarExtensions = []string{".npz"}

var viewerCommand = &cli.Command{
	Name:            "viewer",
	Usage:           "Open local Boltz outputs in the web molecular viewer",
	Suggest:         true,
	HideHelpCommand: true,
	Commands: []*cli.Command{
		viewerOpenCommand,
	},
}

var viewerOpenCommand = &cli.Command{
	Name:            "open",
	Usage:           "Open a local Boltz output directory in the web viewer",
	Suggest:         true,
	HideHelpCommand: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "dir",
			Usage: "Directory containing Boltz output files. A single positional directory may also be used.",
			Value: ".",
		},
		&cli.StringFlag{
			Name:  "workspace-dir",
			Usage: "Parent folder users should grant in the browser to browse sibling design outputs. Defaults to the parent of --dir.",
		},
		&cli.IntFlag{
			Name:  "port",
			Usage: "Localhost port for the temporary file server. 0 chooses a random free port.",
			Value: 0,
		},
		&cli.StringFlag{
			Name:    "viewer-url",
			Usage:   "Viewer page to open.",
			Value:   viewerOpenDefaultURL,
			Sources: cli.EnvVars("BOLTZ_VIEWER_URL"),
		},
		&cli.BoolFlag{
			Name:  "no-open",
			Usage: "Print the viewer URL without launching a browser.",
		},
	},
	Action: handleViewerOpen,
}

type viewerOpenSpec struct {
	RootDir      string
	WorkspaceDir string
	Port         int
	ViewerURL    string
	NoOpen       bool
}

type viewerOpenServer struct {
	rootDir      string
	workspace    viewerOpenManifestWorkspace
	files        []viewerOpenManifestFile
	filePaths    map[string]string
	token        string
	baseURL      string
	manifestURL  string
	viewerURL    string
	viewerOrigin string
	listener     net.Listener
	server       *http.Server
}

type viewerOpenManifest struct {
	SchemaVersion int                         `json:"schema_version"`
	Name          string                      `json:"name"`
	Workspace     viewerOpenManifestWorkspace `json:"workspace"`
	Files         []viewerOpenManifestFile    `json:"files"`
}

type viewerOpenManifestWorkspace struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type viewerOpenManifestFile struct {
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	URL          string `json:"url"`
}

func handleViewerOpen(ctx context.Context, cmd *cli.Command) error {
	spec, err := parseViewerOpenSpec(cmd)
	if err != nil {
		return err
	}

	server, err := startViewerOpenServer(spec)
	if err != nil {
		return err
	}
	defer server.Close()

	if !spec.NoOpen {
		if err := openExternalBrowser(server.viewerURL); err != nil {
			fmt.Fprintf(commandErrorWriter(cmd), "Could not open browser automatically: %v\n", err)
		}
	}

	fmt.Fprintf(commandWriter(cmd), "Viewer URL: %s\n", server.viewerURL)
	fmt.Fprintf(commandWriter(cmd), "Serving %s. Press Ctrl-C to stop.\n", spec.RootDir)

	waitCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-waitCtx.Done()
	return nil
}

func parseViewerOpenSpec(cmd *cli.Command) (viewerOpenSpec, error) {
	args := cmd.Args().Slice()
	if len(args) > 1 {
		return viewerOpenSpec{}, fmt.Errorf("Unexpected extra arguments: %v", args[1:])
	}
	if len(args) == 1 && cmd.IsSet("dir") {
		return viewerOpenSpec{}, errors.New("Use either --dir or a positional directory, not both")
	}

	rootDir := cmd.String("dir")
	if len(args) == 1 {
		rootDir = args[0]
	}
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return viewerOpenSpec{}, err
	}
	rootInfo, err := os.Stat(rootDir)
	if err != nil {
		return viewerOpenSpec{}, err
	}
	if !rootInfo.IsDir() {
		return viewerOpenSpec{}, fmt.Errorf("%s is not a directory", rootDir)
	}

	workspaceDir := cmd.String("workspace-dir")
	if workspaceDir == "" {
		workspaceDir = filepath.Dir(rootDir)
	}
	workspaceDir, err = filepath.Abs(workspaceDir)
	if err != nil {
		return viewerOpenSpec{}, err
	}
	workspaceInfo, err := os.Stat(workspaceDir)
	if err != nil {
		return viewerOpenSpec{}, err
	}
	if !workspaceInfo.IsDir() {
		return viewerOpenSpec{}, fmt.Errorf("%s is not a directory", workspaceDir)
	}

	viewerURL := cmd.String("viewer-url")
	parsedViewerURL, err := neturl.Parse(viewerURL)
	if err != nil {
		return viewerOpenSpec{}, err
	}
	if parsedViewerURL.Scheme != "http" && parsedViewerURL.Scheme != "https" {
		return viewerOpenSpec{}, errors.New("--viewer-url must use http or https")
	}

	return viewerOpenSpec{
		RootDir:      rootDir,
		WorkspaceDir: workspaceDir,
		Port:         cmd.Int("port"),
		ViewerURL:    viewerURL,
		NoOpen:       cmd.Bool("no-open"),
	}, nil
}

func startViewerOpenServer(spec viewerOpenSpec) (*viewerOpenServer, error) {
	token, err := viewerOpenRandomToken()
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viewerOpenLocalhostAddress, spec.Port))
	if err != nil {
		return nil, err
	}

	baseURL := "http://" + listener.Addr().String()
	manifestURL := fmt.Sprintf("%s/t/%s/manifest.json", baseURL, token)
	workspace := viewerOpenWorkspace(spec.WorkspaceDir)
	files, filePaths, err := scanViewerOpenFiles(spec.RootDir, baseURL, token)
	if err != nil {
		listener.Close()
		return nil, err
	}

	viewerURL, viewerOrigin, err := buildViewerOpenURL(spec.ViewerURL, manifestURL, workspace)
	if err != nil {
		listener.Close()
		return nil, err
	}

	viewServer := &viewerOpenServer{
		rootDir:      spec.RootDir,
		workspace:    workspace,
		files:        files,
		filePaths:    filePaths,
		token:        token,
		baseURL:      baseURL,
		manifestURL:  manifestURL,
		viewerURL:    viewerURL,
		viewerOrigin: viewerOrigin,
		listener:     listener,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/t/"+token+"/manifest.json", viewServer.handleManifest)
	mux.HandleFunc("/t/"+token+"/files/", viewServer.handleFile)
	viewServer.server = &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		if err := viewServer.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "viewer server stopped: %v\n", err)
		}
	}()

	return viewServer, nil
}

func (server *viewerOpenServer) Close() {
	if server.server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.server.Shutdown(shutdownCtx)
	}
	if server.listener != nil {
		_ = server.listener.Close()
	}
}

func (server *viewerOpenServer) handleManifest(w http.ResponseWriter, r *http.Request) {
	if server.handleCORS(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(viewerOpenManifest{
		SchemaVersion: viewerOpenManifestVersion,
		Name:          filepath.Base(server.rootDir),
		Workspace:     server.workspace,
		Files:         server.files,
	})
}

func (server *viewerOpenServer) handleFile(w http.ResponseWriter, r *http.Request) {
	if server.handleCORS(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	relativePath := strings.TrimPrefix(r.URL.Path, "/t/"+server.token+"/files/")
	filePath, err := server.resolveFilePath(relativePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if strings.HasSuffix(strings.ToLower(filePath), ".cif.gz") {
		w.Header().Set("Content-Type", "chemical/x-cif")
		w.Header().Set("Content-Encoding", "gzip")
	} else if contentType := viewerOpenContentType(filePath); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filepath.Base(filePath)))
	http.ServeFile(w, r, filePath)
}

func (server *viewerOpenServer) handleCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin != "" && server.isAllowedOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Private-Network", "true")
		w.Header().Set("Vary", "Origin")
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func (server *viewerOpenServer) isAllowedOrigin(origin string) bool {
	if origin == server.viewerOrigin {
		return true
	}
	parsed, err := neturl.Parse(origin)
	if err != nil {
		return false
	}
	return (parsed.Scheme == "http" || parsed.Scheme == "https") &&
		slices.Contains([]string{"localhost", "127.0.0.1", "::1"}, parsed.Hostname())
}

func (server *viewerOpenServer) resolveFilePath(relativePath string) (string, error) {
	cleanPath := strings.TrimPrefix(path.Clean("/"+relativePath), "/")
	if cleanPath == "." || cleanPath == "" {
		return "", errors.New("missing file path")
	}

	filePath, ok := server.filePaths[cleanPath]
	if !ok {
		return "", errors.New("file not found in viewer manifest")
	}
	return filePath, nil
}

func scanViewerOpenFiles(rootDir string, baseURL string, token string) ([]viewerOpenManifestFile, map[string]string, error) {
	var files []viewerOpenManifestFile
	filePaths := map[string]string{}
	err := filepath.WalkDir(rootDir, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		kind := viewerOpenFileKind(filePath)
		if kind == "" {
			return nil
		}
		relativePath, err := filepath.Rel(rootDir, filePath)
		if err != nil {
			return err
		}
		relativePath = filepath.ToSlash(relativePath)
		files = append(files, viewerOpenManifestFile{
			Kind:         kind,
			Name:         filepath.Base(filePath),
			Path:         relativePath,
			Size:         info.Size(),
			LastModified: info.ModTime().UTC().Format(time.RFC3339),
			URL:          fmt.Sprintf("%s/t/%s/files/%s", baseURL, token, viewerOpenEscapePath(relativePath)),
		})
		filePaths[relativePath] = filePath
		return nil
	})
	return files, filePaths, err
}

func viewerOpenFileKind(filePath string) string {
	lowerPath := strings.ToLower(filePath)
	if slices.ContainsFunc(viewerOpenStructureExtensions, func(extension string) bool {
		return strings.HasSuffix(lowerPath, extension)
	}) {
		return "structure"
	}
	if slices.ContainsFunc(viewerOpenSidecarExtensions, func(extension string) bool {
		return strings.HasSuffix(lowerPath, extension)
	}) {
		return "sidecar"
	}
	return ""
}

func viewerOpenEscapePath(relativePath string) string {
	parts := strings.Split(filepath.ToSlash(relativePath), "/")
	for index, part := range parts {
		parts[index] = neturl.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func viewerOpenContentType(filePath string) string {
	lowerPath := strings.ToLower(filePath)
	switch {
	case strings.HasSuffix(lowerPath, ".cif"):
		return "chemical/x-cif"
	case strings.HasSuffix(lowerPath, ".pdb"):
		return "chemical/x-pdb"
	case strings.HasSuffix(lowerPath, ".bcif"), strings.HasSuffix(lowerPath, ".npz"):
		return "application/octet-stream"
	default:
		return mime.TypeByExtension(filepath.Ext(filePath))
	}
}

func viewerOpenWorkspace(workspaceDir string) viewerOpenManifestWorkspace {
	hash := sha256.Sum256([]byte(workspaceDir))
	return viewerOpenManifestWorkspace{
		Key:  hex.EncodeToString(hash[:16]),
		Name: filepath.Base(workspaceDir),
	}
}

func buildViewerOpenURL(viewerURL string, manifestURL string, workspace viewerOpenManifestWorkspace) (string, string, error) {
	parsedViewerURL, err := neturl.Parse(viewerURL)
	if err != nil {
		return "", "", err
	}
	query := parsedViewerURL.Query()
	query.Set("manifestUrl", manifestURL)
	query.Set("workspaceKey", workspace.Key)
	query.Set("workspaceName", workspace.Name)
	parsedViewerURL.RawQuery = query.Encode()
	return parsedViewerURL.String(), parsedViewerURL.Scheme + "://" + parsedViewerURL.Host, nil
}

func viewerOpenRandomToken() (string, error) {
	raw := make([]byte, viewerOpenTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func openExternalBrowser(target string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", target)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		command = exec.Command("xdg-open", target)
	}
	return command.Start()
}
