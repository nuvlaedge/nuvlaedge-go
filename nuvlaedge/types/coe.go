package types

type LegacyJobConf struct {
	Image            string
	ApiKey           string
	ApiSecret        string
	Endpoint         string
	EndpointInsecure bool
	JobId            string
	NuvlaedgeFsPath  string
}
