package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	log "github.com/sirupsen/logrus"
	"slices"
	"sync"
	"time"
)

type gatherFunc struct {
	needSwarm    bool
	resourceName string
	retrieveFunc func(ctx context.Context) (interface{}, error)
	dest         *[]map[string]interface{}
}

func (dm *DockerMonitor) getGatherers() []gatherFunc {
	gatherers := []gatherFunc{
		{
			false,
			"images",
			func(ctx context.Context) (interface{}, error) { return dm.client.ImageList(ctx, image.ListOptions{}) },
			&dm.coeResources.DockerResources.Images,
		},
		{
			false,
			"containers",
			func(ctx context.Context) (interface{}, error) {
				return dm.client.ContainerList(ctx, container.ListOptions{})
			},
			&dm.coeResources.DockerResources.Containers,
		},
		{
			false,
			"volumes",
			func(ctx context.Context) (interface{}, error) {
				v, err := dm.client.VolumeList(ctx, volume.ListOptions{})
				return v.Volumes, err
			},
			&dm.coeResources.DockerResources.Volumes,
		},
		{
			false,
			"networks",
			func(ctx context.Context) (interface{}, error) {
				return dm.client.NetworkList(ctx, network.ListOptions{})
			},
			&dm.coeResources.DockerResources.Networks,
		},
	}

	_, err := dm.client.SwarmInspect(context.Background())
	if err != nil {
		log.Warn("Swarm not found, skipping swarm resources: ", err)
		return gatherers
	}

	swarmGatherers := []gatherFunc{
		{
			true,
			"services",
			func(ctx context.Context) (interface{}, error) {
				return dm.client.ServiceList(ctx, types.ServiceListOptions{})
			},
			&dm.coeResources.DockerResources.Services,
		},
		{
			true,
			"tasks",
			func(ctx context.Context) (interface{}, error) {
				return dm.client.TaskList(ctx, types.TaskListOptions{})
			},
			&dm.coeResources.DockerResources.Tasks},
		{
			true,
			"configs",
			func(ctx context.Context) (interface{}, error) {
				return dm.client.ConfigList(ctx, types.ConfigListOptions{})
			},
			&dm.coeResources.DockerResources.Configs},
		{
			true,
			"secrets",
			func(ctx context.Context) (interface{}, error) {
				return dm.client.SecretList(ctx, types.SecretListOptions{})
			},
			&dm.coeResources.DockerResources.Secrets},
	}

	return slices.Concat(gatherers, swarmGatherers)
}

func (dm *DockerMonitor) updateCoeResources() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	gatherers := dm.getGatherers()
	wg := sync.WaitGroup{}
	wg.Add(len(gatherers))
	var errs []error
	errMutex := sync.Mutex{}

	for _, g := range gatherers {
		go func(g gatherFunc) {
			defer wg.Done()
			resources, err := dm.retrieveResources(ctx, g.retrieveFunc)
			if err != nil {
				errMutex.Lock()
				errs = append(errs, fmt.Errorf("error retrieving %s: %s", g.resourceName, err))
				errMutex.Unlock()
				return
			}
			*g.dest = resources
		}(g)
	}

	wg.Wait()

	return errors.Join(errs...)
}

func (dm *DockerMonitor) retrieveResources(ctx context.Context, retrieveFunc func(context.Context) (interface{}, error)) ([]map[string]interface{}, error) {
	resources, err := retrieveFunc(ctx)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	b, err := json.Marshal(resources)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
