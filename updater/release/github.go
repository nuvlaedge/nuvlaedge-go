package release

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"nuvlaedge-go/updater/common"
	"path/filepath"
	"slices"
)

const (
	// GitHubReleaseURL is the URL of the GitHub API to get the releases
	GitHubReleaseURL     = "https://github.com/nuvlaedge/nuvlaedge-go/releases"
	GitHubReleasesAPIURL = "https://api.github.com/repos/nuvlaedge/nuvlaedge-go/releases"
)

type GitHubAsset struct {
	Url                string `json:"url"`
	Id                 string `json:"id"`
	Name               string `json:"name"`
	Label              string `json:"label"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type GitHubRelease struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`

	TagName    string        `json:"tag_name"`
	Name       string        `json:"name"`
	Assets     []GitHubAsset `json:"assets"`
	PreRelease bool          `json:"prerelease"`
}

func (gt *GitHubRelease) GetComposeFiles(fileNames []string, workDir string) ([]string, error) {
	var files []string
	for _, asset := range gt.Assets {
		if slices.Contains(fileNames, asset.Name) {
			err := common.DownloadFile(asset.Url, filepath.Join(workDir, asset.Name))
			if err != nil {
				log.Errorf("Error downloading asset %s: %s", asset.Name, err)
				return nil, err
			}
		}
	}

	return files, nil
}

func GetGitHubRelease(version string) (*GitHubRelease, error) {
	releases, err := ListGitHubReleases()
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, errors.New("no releases found")
	}

	for _, release := range releases {
		if release.TagName == version {
			return &release, nil
		}
	}

	log.Infof("No matching release found for version %s, returning latest and error", version)
	return &releases[0], errors.New("no release found")
}

func ListGitHubReleases() ([]GitHubRelease, error) {
	resp, err := http.Get(GitHubReleasesAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	var r []GitHubRelease
	err = json.NewDecoder(resp.Body).Decode(&r)

	return r, nil
}
