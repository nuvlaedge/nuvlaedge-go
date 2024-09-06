package release

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"nuvlaedge-go/updater/common"
	"path/filepath"
	"slices"
)

type NuvlaComposeFile struct {
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	FileContent string `json:"file"`
}

type NuvlaReleaseResource struct {
	ReleaseDate  string             `json:"release-date"`
	PreRelease   bool               `json:"prerelease"`
	ComposeFiles []NuvlaComposeFile `json:"compose-files"`
	Id           string             `json:"id"`
	Url          string             `json:"url"`
	Release      string             `json:"release"`
	Published    bool               `json:"published"`
}

func (nr *NuvlaReleaseResource) GetComposeFiles(fileNames []string, workDir string) ([]string, error) {
	var composeFiles []string
	for _, composeFile := range nr.ComposeFiles {
		if slices.Contains(fileNames, composeFile.Name) {
			// Save file to disk
			err := common.SaveFile(composeFile.Name, workDir, composeFile.FileContent)
			if err != nil {
				log.Errorf("Error saving compose file %s: %s", composeFile.Name, err)
				return nil, err
			}
			composeFiles = append(composeFiles, filepath.Join(workDir, composeFile.Name))
		}
	}

	if len(composeFiles) == 0 {
		return nil, errors.New("no compose files found")
	}

	return composeFiles, nil
}

const (
	NuvlaEndpoint = "https://nuvla.io/api/nuvlabox-release"
)

func GetNuvlaRelease(version string) (*NuvlaReleaseResource, error) {
	ret, err := http.Get(NuvlaEndpoint)
	if err != nil {
		return nil, err
	}
	var res struct {
		Releases []NuvlaReleaseResource `json:"resources"`
		Id       string                 `json:"id"`
	}

	defer ret.Body.Close()

	err = json.NewDecoder(ret.Body).Decode(&res)
	if err != nil {
		return nil, err
	}

	rel := res.Releases

	if len(rel) == 0 {
		log.Warnf("No nuvla releases found")
		return nil, errors.New("no nuvla releases found")
	}

	log.Infof("Found %d nuvla releases", len(rel))
	if version == "" || version == "latest" {
		return &rel[len(rel)-1], nil
	}

	for _, release := range rel {
		if release.Release == version {
			return &release, nil
		}
	}

	// Return the latest element if the version is not found
	log.Info("Version not found, returning the latest release")
	return &rel[len(rel)-1], errors.New("requested version not found")
}
