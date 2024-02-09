package types

import "fmt"

type NuvlaEdgeComponent interface {
	GetNuvlaEdgeComponentType() NuvlaEdgeComponentType
	GetNuvlaEdgeComponentVersion() (string, error)
	GetNuvlaEdgeComponentID() string
	String() string
}

type NuvlaEdgeComponentType string

const (
	ContainerType NuvlaEdgeComponentType = "container"
	ProcessType   NuvlaEdgeComponentType = "process"
)

type NuvlaEdgeContainer struct {
	componentType NuvlaEdgeComponentType
	containerID   string
	nameId        string
	version       string
}

func NewNuvlaEdgeContainer(containerID string, containerName string) *NuvlaEdgeContainer {
	return &NuvlaEdgeContainer{
		componentType: ContainerType,
		containerID:   containerID,
		nameId:        containerName,
	}
}

// Implement NuvlaEdgeComponent interface for NuvlaEdgeContainer

// GetNuvlaEdgeComponentType returns the container ID
func (c *NuvlaEdgeContainer) GetNuvlaEdgeComponentType() NuvlaEdgeComponentType {
	return c.componentType
}

// GetNuvlaEdgeComponentVersion returns the container version
func (c *NuvlaEdgeContainer) GetNuvlaEdgeComponentVersion() (string, error) {
	return c.version, nil
}

// GetNuvlaEdgeComponentID returns the container ID
func (c *NuvlaEdgeContainer) GetNuvlaEdgeComponentID() string {
	return c.containerID
}

// String returns the container ID
func (c *NuvlaEdgeContainer) String() string {
	return fmt.Sprintf("NuvlaEdge component, type %s, ID %s", c.componentType, c.containerID)
}
