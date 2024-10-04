package testutils

import (
	"context"
	nuvlaTypes "github.com/nuvla/api-client-go/types"
)

type CommissionerClientMock struct {
	GetCnt    int
	GetReturn nuvlaTypes.NuvlaResource
	GetError  error

	CommissionCnt int
	CommissionErr error

	GetStatusIdCnt    int
	GetStatusIdReturn string
}

func (c *CommissionerClientMock) Get(ctx context.Context, id string, selectFields []string) (*nuvlaTypes.NuvlaResource, error) {
	c.GetCnt++
	return &c.GetReturn, c.GetError
}

func (c *CommissionerClientMock) Commission(ctx context.Context, data map[string]interface{}) error {
	c.CommissionCnt++
	return c.CommissionErr
}

func (c *CommissionerClientMock) GetStatusId() string {
	c.GetStatusIdCnt++
	return c.GetStatusIdReturn
}
