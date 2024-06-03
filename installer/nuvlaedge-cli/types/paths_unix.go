package types

import (
	"fmt"
	"nuvlaedge-cli/common"
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
		bPath, _ = filepath.Abs(common.BinarySudoPath)
		cPath, _ = filepath.Abs(common.ConfigSudoPath)
		lPath, _ = filepath.Abs(common.LogsSudoPath)
		dbPath, _ = filepath.Abs(common.DatabaseSudoPath)
	} else {
		usr, _ := user.Current()
		bPath = filepath.Join(usr.HomeDir, common.BinaryPath)
		cPath = filepath.Join(usr.HomeDir, common.ConfigPath)
		lPath = filepath.Join(usr.HomeDir, common.LogsPath)
		dbPath = filepath.Join(usr.HomeDir, common.DatabasePath)
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
