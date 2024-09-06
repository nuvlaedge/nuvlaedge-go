package util

import (
	"encoding/json"
	"errors"
	"os"
)

type NuvlaEdgeInfo struct {
	NuvlaEdgeUUID string `json:"nuvlaedge-uuid"`
	ApiKey        string `json:"key"`
	ApiSecret     string `json:"secret"`
}

func GetNuvlaEdgeInfo(filePath string) (*NuvlaEdgeInfo, error) {
	info, err := getInfoFromEnv()
	if err == nil {
		return info, nil
	}

	info, err2 := getInfoFromFile(filePath)
	if err2 == nil {
		return info, nil
	}

	return nil, errors.Join(err, err2)
}

func getInfoFromEnv() (*NuvlaEdgeInfo, error) {
	id := os.Getenv("NUVLAEDGE_UUID")
	key := os.Getenv("NUVLAEDGE_API_KEY")
	secret := os.Getenv("NUVLAEDGE_API_SECRET")
	if id == "" || key == "" || secret == "" {
		return nil, errors.New("missing NuvlaEdge info in environment")
	}
	return &NuvlaEdgeInfo{
		NuvlaEdgeUUID: id,
		ApiKey:        key,
		ApiSecret:     secret,
	}, nil
}

func getInfoFromFile(filePath string) (*NuvlaEdgeInfo, error) {
	// Open file and read content
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New("failed to open file: " + err.Error())
	}
	defer file.Close()

	// Read the file content
	var info NuvlaEdgeInfo
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&info); err != nil {
		return nil, errors.New("failed to decode JSON: " + err.Error())
	}
	return &info, nil
}
