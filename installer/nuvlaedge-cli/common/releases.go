package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func (r Release) String() string {
	return r.TagName
}

// getAllReleases fetches all releases from a GitHub repository given the owner and the repository name
func getAllReleases(owner string, repo string) ([]Release, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo))
	if err != nil {
		fmt.Println("Error fetching releases")
		return nil, err
	}
	defer resp.Body.Close()

	var releases []Release
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body")
		return nil, err
	}

	err = json.Unmarshal(body, &releases)
	if err != nil {
		fmt.Println("Error unmarshaling response body")
		return nil, err
	}

	return releases, nil
}

// GetVersion asserts the version exists in the repository and returns the version to be installed.
// It accepts "latest" as a version. And empty string is also accepted, and it will return the latest version.
func GetVersion(v, org, repo string) string {
	// 1. Get all available versions
	releases, err := getAllReleases(org, repo)
	if err != nil {
		fmt.Println("Error getting releases")
		return ""
	}
	fmt.Println("Releases: ", releases)

	// 2.1 If v is "latest", return the latest version
	if v == "latest" {
		return releases[0].TagName
	}

	// 2.2 If v is not "latest", return check availability and return the version
	for _, r := range releases {
		cleanRemote := strings.Replace(r.TagName, "v", "", 1)
		cleanLocal := strings.Replace(v, "v", "", 1)
		if cleanLocal == cleanRemote {
			return v
		}
	}
	fmt.Printf("Requested version %s not found, installing latest\n", v)
	return releases[0].TagName
}
