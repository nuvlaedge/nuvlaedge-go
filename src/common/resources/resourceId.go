package resources

import (
	"fmt"
	"strings"
)

type ResourceId struct {
	Id   string
	Type string
	Uuid string
}

func NewResourceId(id string) (*ResourceId, error) {
	sep := strings.Split(id, "/")
	if len(sep) < 2 {
		return nil, fmt.Errorf("provided resource ID does not fulfil the Nuvla id standard <resType>/<uuid> ")
	}
	return &ResourceId{
		Id:   id,
		Type: sep[0],
		Uuid: sep[1],
	}, nil
}

func (r *ResourceId) toString() string {
	return r.Id
}

func (r *ResourceId) String() string {
	return r.toString()
}
