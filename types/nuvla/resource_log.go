package nuvla

import (
	"encoding/json"
	"errors"
	nuvlaapi "github.com/nuvla/api-client-go"
)

type Logs map[string][]string

type ResourceLogResource struct {
	Id            string   `json:"id"`
	ResourceType  string   `json:"resource-type"`
	Parent        string   `json:"parent"`
	Components    []string `json:"components"`
	Since         string   `json:"since"`
	LastTimeStamp string   `json:"last-timestamp"`
	Lines         int16    `json:"lines"`
	Log           Logs     `json:"log"`
}

type ResourceLogClient struct {
	*nuvlaapi.NuvlaClient

	logResource   *ResourceLogResource
	ResourceLogId string
}

func NewResourceLogClient(client *nuvlaapi.NuvlaClient) *ResourceLogClient {
	return &ResourceLogClient{
		NuvlaClient: client,
	}
}

func (rlc *ResourceLogClient) UpdateResourceSelect(selects []string) error {
	res, err := rlc.Get(rlc.ResourceLogId, selects)
	if err != nil {
		return err
	}

	if rlc.logResource == nil {
		rlc.logResource = &ResourceLogResource{}
	}

	b, err := json.Marshal(res.Data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, rlc.logResource)
	if err != nil {
		return err
	}

	return nil
}

func (rlc *ResourceLogClient) UpdateResource() error {
	return rlc.UpdateResourceSelect(nil)
}

func (rlc *ResourceLogClient) getSince() string {
	if rlc.logResource.LastTimeStamp == "" {
		return rlc.logResource.Since
	}
	return rlc.logResource.LastTimeStamp
}

func (rlc *ResourceLogClient) UpdateLogs(logs *Logs) error {
	//rlc.Edit(rlc.ResourceLogId, logs, nil)
	if logs == nil || len(*logs) == 0 {
		return errors.New("no logs to update")
	}

	resource := ResourceLogResource{}
	resource.Log = *logs

	return nil
}
