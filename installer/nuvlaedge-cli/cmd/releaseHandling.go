package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func (r Release) String() string {
	return r.TagName
}

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
