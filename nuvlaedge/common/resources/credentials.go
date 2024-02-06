package resources

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/nuvlaedge/common"
	"os"
)

type Credentials struct {
	Key    string `mapstructure:"key" json:"api-key"`
	Secret string `mapstructure:"secret" json:"secret-key"`
	Href   string `mapstructure:"href" json:"href"`
}

type CredentialPayload = map[string]string

func NewCredentialsFromBody(body []byte) *Credentials {
	c := Credentials{}
	err := json.Unmarshal(body, &c)
	common.GenericErrorHandler("Error unmarshalling body into credentials", err)

	// Define constant value
	c.Href = common.SessionTemplate
	return &c
}

func (c *Credentials) FormatCredentialsPayload() CredentialPayload {
	return CredentialPayload{
		"key":    c.Key,
		"secret": c.Secret,
		"href":   c.Href,
	}
}

func (c *Credentials) ToString() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func (c *Credentials) Dump(file string) {
	f, err := os.Create(file)
	if err != nil {
		// failed to create/open the file
		log.Fatal(err)
	}
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		// failed to encode
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		// failed to close the file
		log.Fatal(err)

	}
}

func NewFromToml(fileName string) (*Credentials, error) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	var creds Credentials
	decoder := toml.NewDecoder(file)
	_, err = decoder.Decode(&creds)
	if err != nil {
		return nil, err
	}

	fmt.Println("Credentials loaded from file:")
	fmt.Println("Key:", creds.Key)
	fmt.Println("Secret:", creds.Secret)
	fmt.Println("Href:", creds.Href)

	return &creds, nil
}
