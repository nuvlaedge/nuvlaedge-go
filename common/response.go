package common

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type NEResponse struct {
	Jobs       []string `json:"jobs"`
	LastUpdate string   `json:"doc-last-updated"`
}

func NewFromResponse(res *http.Response) (*NEResponse, error) {
	var neRes NEResponse
	err := json.NewDecoder(res.Body).Decode(&neRes)
	if err != nil {
		log.Error("Error decoding NuvlaEdge response: ", err)
		return nil, err
	}

	return &neRes, nil
}

func ProcessResponse(res *http.Response, jobChan chan string, confChan chan string) error {
	neRes, err := NewFromResponse(res)
	if err != nil {
		return err
	}

	if neRes.Jobs != nil && len(neRes.Jobs) > 0 {
		log.Infof("Received %d jobs", len(neRes.Jobs))
		for _, job := range neRes.Jobs {
			jobChan <- job
		}
	}

	if neRes.LastUpdate != "" {
		log.Infof("Received last update: %s", neRes.LastUpdate)
		confChan <- neRes.LastUpdate
	}

	return nil
}
