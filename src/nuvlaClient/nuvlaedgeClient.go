package nuvlaClient

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"native-nuvlaedge/src/agent/commissioner"
	"native-nuvlaedge/src/common"
	"native-nuvlaedge/src/common/resources"
	"net/http"
)

type ReportResponse struct {
	Jobs []string `json:"jobs"`
}

type NuvlaEdgeClient struct {
	Uuid                string
	NuvlaEdgeStatusUuid string
	StatusUuid          string
	client              *NuvlaClient
	Credentials         *resources.Credentials
	NuvlaEdgeResource   *resources.NuvlaEdge
}

func NewNuvlaEdgeClient(uuid string, endpoint string, insecure bool) *NuvlaEdgeClient {
	return &NuvlaEdgeClient{
		Uuid:   uuid,
		client: NewNuvlaClient(endpoint, insecure),
	}
}

func (nec *NuvlaEdgeClient) Activate() error {
	if common.FileExists("/tmp/sample_creds.toml") {
		preloadCreds, _ := resources.NewFromToml("/tmp/sample_creds.toml")
		log.Infof("Credentials found: %s", preloadCreds)
		nec.Credentials = preloadCreds
		return nil
	}
	log.Info("No credentials found, assuming NuvlaEdge NEW state")

	endPointPath := "/api/" + nec.Uuid + "/activate"
	emptyPayload := map[string]any{}
	resp, err := nec.client.Post(emptyPayload, endPointPath)
	common.GenericErrorHandler("Error in Activate Post", err)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error activating with status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	common.GenericErrorHandler("error reading activate response", err)

	nec.Credentials = resources.NewCredentialsFromBody(body)
	log.Debugf("Saved NuvlaEdge credentials %s:", nec.Credentials.ToString())
	nec.Credentials.Dump("/tmp/sample_creds.toml")
	return nil
}

func (nec *NuvlaEdgeClient) GetNuvlaEdgeInformation() error {

	endpoint := "/api/" + nec.Uuid

	resp, err := nec.client.Get(endpoint)

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Errorf("Cannot retrieve NuvlaEdge info for %s, exiting", nec.Uuid)
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	common.GenericErrorHandler("error reading nuvlabox retrieve response", err)

	nuvlaEdgeRes := resources.NuvlaEdge{}
	err = json.Unmarshal(body, &nuvlaEdgeRes)
	common.GenericErrorHandler("error unmarshalling nuvlabox resource", err)

	if nec.NuvlaEdgeResource == nil {
		nec.NuvlaEdgeResource = &nuvlaEdgeRes
		nec.NuvlaEdgeStatusUuid = nuvlaEdgeRes.NuvlaEdgeStatusId
	}
	return nil
}

// Commission executes the commissioning action of the given agent against Nuvla.
func (nec *NuvlaEdgeClient) Commission(data *commissioner.CommissioningAttributes) (bool, error) {

	if !nec.client.Authenticated() {
		nec.client.LogIn(nec.Credentials)
	}

	commissionPayload := map[string]any{
		"capabilities": []string{"NUVLA_JOB_PULL", "NUVLA_HEARTBEAT"},
		"status":       "UNKNOWN"}

	log.Info("I should now be authenticated")

	log.Info("Running commissioning process")
	endPointPath := "/api/" + nec.Uuid + "/commission"
	resp, err := nec.client.Post(commissionPayload, endPointPath)
	common.GenericErrorHandler("Error commissioning", err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	common.GenericErrorHandler("error reading activate response", err)

	log.Info(string(body))

	return true, nil
}

func (nec *NuvlaEdgeClient) HeartBeat() ([]string, error) {
	log.Info("Sending Heartbeat")
	endPointPath := "/api/" + nec.Uuid + "/heartbeat"
	resp, err := nec.client.Post(nil, endPointPath)
	common.GenericErrorHandler("Error commissioning", err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	common.GenericErrorHandler("error reading activate response", err)

	jobs := ReportResponse{}
	err = json.Unmarshal(body, &jobs)

	if err != nil {
		log.Errorf("error unmarshaling %s", err)
		return []string{}, nil
	}

	return jobs.Jobs, nil
}

func (nec *NuvlaEdgeClient) Telemetry(info *resources.NuvlaEdgeStatus) ([]string, error) {
	log.Info("Sending Telemetry")
	endPoint := "/api/" + nec.NuvlaEdgeStatusUuid

	resp, err := nec.client.Put(info, endPoint)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	common.GenericErrorHandler("error reading activate response", err)

	jobs := ReportResponse{}
	err = json.Unmarshal(body, &jobs)

	if err != nil {
		log.Errorf("error unmarshaling %s", err)
		return []string{}, nil
	}

	return jobs.Jobs, nil
}
