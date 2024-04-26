package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
)

type InstallPaths struct {
	BinaryPath   string
	ConfigPath   string
	LogsPath     string
	DatabasePath string
}

func NewFromBasePath(basePath string) InstallPaths {
	bPath, _ := filepath.Abs(fmt.Sprintf("%s/bin/", basePath))
	cPath, _ := filepath.Abs(fmt.Sprintf("%s/config/", basePath))
	lPath, _ := filepath.Abs(fmt.Sprintf("%s/logs/", basePath))
	dbPath, _ := filepath.Abs(fmt.Sprintf("%s/db/", basePath))

	return InstallPaths{
		BinaryPath:   bPath,
		ConfigPath:   cPath,
		LogsPath:     lPath,
		DatabasePath: dbPath,
	}
}

func NewDefaultInstallPaths(isSudo bool) InstallPaths {
	var bPath string
	var cPath string
	var lPath string
	var dbPath string

	if isSudo {
		bPath, _ = filepath.Abs(BinarySudoPath)
		cPath, _ = filepath.Abs(ConfigSudoPath)
		lPath, _ = filepath.Abs(LogsSudoPath)
		dbPath, _ = filepath.Abs(DatabaseSudoPath)
	} else {
		usr, _ := user.Current()
		bPath = filepath.Join(usr.HomeDir, BinaryPath)
		cPath = filepath.Join(usr.HomeDir, ConfigPath)
		lPath = filepath.Join(usr.HomeDir, LogsPath)
		dbPath = filepath.Join(usr.HomeDir, DatabasePath)
	}

	return InstallPaths{
		BinaryPath:   bPath,
		ConfigPath:   cPath,
		LogsPath:     lPath,
		DatabasePath: dbPath,
	}
}

func (p InstallPaths) MakePaths() {
	v := reflect.ValueOf(p)
	t := reflect.TypeOf(p)

	for i := 0; i < v.NumField(); i++ {
		path := v.Field(i).Interface().(string)
		fmt.Printf("Creating %s path: %s\n", t.Field(i).Name, path)
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating %s path: %v\n", t.Field(i).Name, err)
			panic(err)
		}
	}
}

func (p InstallPaths) String() string {
	return fmt.Sprintf("BinaryPath: %s, ConfigPath: %s", p.BinaryPath, p.ConfigPath)
}

// URL constants
const (
	// NuvlaEdgeBinaryURL NuvlaEdge binary url
	NuvlaEdgeBinaryURL = "https://github.com/nuvlaedge/nuvlaedge-go/releases/download/%s/nuvlaedge-%s-%s-%s"
	// NuvlaEdgeLatestReleaseURL NuvlaEdge latest release url
	NuvlaEdgeLatestReleaseURL = "https://api.github.com/repos/nuvlaedge/nuvlaedge-go/releases/latest"
	// NuvlaEdgeLatestConfTemplateURL NuvlaEdge latest configuration template url
	NuvlaEdgeLatestConfTemplateURL = "https://raw.githubusercontent.com/nuvlaedge/nuvlaedge-go/main/config/template.toml"
	// ReleasesURL NuvlaEdge releases url
	ReleasesURL = "https://api.github.com/repos/{owner}/{repo}/releases"
)

// Path constants
const (
	// BinarySudoPath default path for the binary when running with sudo
	BinarySudoPath = "/usr/local/bin/"
	// ConfigSudoPath default path for the configuration when running with sudo
	ConfigSudoPath = "/etc/nuvlaedge/"
	// DatabaseSudoPath default path for the database when running with sudo
	DatabaseSudoPath = "/var/lib/nuvlaedge/"
	// LogsSudoPath default path for the logs when running with sudo
	LogsSudoPath = "/var/log/nuvlaedge/"

	// BinaryPath default path for the binary
	BinaryPath = ".nuvlaedge/bin/"
	// ConfigPath default path for the configuration
	ConfigPath = ".nuvlaedge/config/"
	// DatabasePath default path for the database
	DatabasePath = ".nuvlaedge/db/"
	// LogsPath default path for the logs
	LogsPath = ".nuvlaedge/logs/"

	// TempDir temporal directory for downloads
	TempDir = "/tmp/nuvlaedge"
)

func hasSudoPermissions() bool {
	return os.Geteuid() == 0
}

func downloadFile(url string, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
