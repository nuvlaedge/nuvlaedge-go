package types

import (
	nuvla "github.com/nuvla/api-client-go"
	"github.com/nuvla/api-client-go/types"
)

type Freeze interface {
	ValidateFileContents(file string) (*Freeze, error)
	LoadFreeze(file string) (*Freeze, error)
}

type NuvlaSessionFreeze struct {
	SessionOpts       *nuvla.SessionOptions    `json:"session-opts"`
	NuvlaEdgeId       string                   `json:"nuvlaedge-id"`
	NuvlaEdgeStatusId string                   `json:"nuvlaedge-status-id"`
	Credentials       *types.ApiKeyLogInParams `json:"credentials"`
}

// ValidateSessionFileContents validates the contents of a file so that it meets the minimum requirements for a
// NuvlaEdge session to be restored using this structure.
func (ns *NuvlaSessionFreeze) ValidateSessionFileContents(file string) (bool, error) {
	return true, nil
}

// LoadSessionFile loads the contents of a file into a NuvlaSessionFreeze structure.
func (ns *NuvlaSessionFreeze) LoadFreeze(file string) error {
	return nil
}

type CommissioningDataFreeze struct {
}
