package types

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type UpdateJobPayload struct {
	ProjectName    string   `json:"project-name"`
	WorkingDir     string   `json:"working-dir"`
	CurrentVersion string   `json:"current-version"`
	Environment    []string `json:"environment"`
	ConfigFiles    []string `json:"config-files"`
	// The target release is the UUID of the Nuvla release to which the NuvlaEdge should be updated
	TargetReleaseUUID string `json:"target-release,omitempty"`
	TargetResource    *NuvlaEdgeReleaseResource
}

func NewPayloadFromString(s string) (*UpdateJobPayload, error) {
	var p UpdateJobPayload
	log.Infof("Parsing job payload")
	err := json.Unmarshal([]byte(s), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type NuvlaEdgeReleaseResource struct {
	Release      string            `json:"release"`
	Url          string            `json:"url"`
	PreRelease   bool              `json:"pre-release"`
	ReleaseDate  string            `json:"release-date"`
	ComposeFiles []ComposeFileInfo `json:"compose-files"`
	ReleaseNotes string            `json:"release-notes"`
	Published    bool              `json:"published"`
}

type ComposeFileInfo struct {
	Name  string `json:"name"`
	File  string `json:"file"`
	Scope string `json:"scope,omitempty"`
}

func NewReleaseResourceFromMap(m map[string]interface{}) (*NuvlaEdgeReleaseResource, error) {
	r := &NuvlaEdgeReleaseResource{}
	b, err := json.Marshal(m)
	if err != nil {

		return nil, err
	}

	err = json.Unmarshal(b, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
