package monitor

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	log "github.com/sirupsen/logrus"
	neTypes "nuvlaedge-go/types"
	"slices"
	"strings"
	"sync"
	"time"
)

type gatherer struct {
	needSwarm    bool
	resourceName string
	retrieveFunc gathererFunc
	dest         *[]map[string]interface{}
}

type gathererFunc func(ctx context.Context, dCli neTypes.DockerMetricsClient) (interface{}, error)

type extendedImage struct {
	image.Summary
	Repository string
	Tag        string
}

func sortImages(ctx context.Context, dCli neTypes.DockerMetricsClient) (interface{}, error) {
	images, err := dCli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(images, func(a, b image.Summary) int {
		if a.Created != b.Created {
			return cmp.Compare(a.Created, b.Created)
		}
		return cmp.Compare(a.ID, b.ID)
	})

	eImages := make([]extendedImage, len(images))
	for i, img := range images {
		slices.Sort(img.RepoTags)
		slices.Sort(img.RepoDigests)
		if len(img.Manifests) > 1 {
			slices.SortFunc(img.Manifests, func(a, b image.ManifestSummary) int {
				return cmp.Compare(a.ID, b.ID)
			})
		}

		eImages[i] = extendedImage{Summary: img}

		if len(img.RepoTags) > 0 {
			repo := strings.SplitN(img.RepoTags[0], ":", 1)
			eImages[i].Repository = repo[0]
			if len(repo) > 1 {
				eImages[i].Tag = repo[1]
			}
		} else if len(img.RepoDigests) > 0 {
			eImages[i].Repository = strings.SplitN(img.RepoDigests[0], "@", 1)[0]
		}
	}

	return eImages, nil
}

type extendedContainer struct {
	types.Container
	Name string
}

func sortContainers(ctx context.Context, dCli neTypes.DockerMetricsClient) (interface{}, error) {
	containers, err := dCli.ContainerList(ctx, container.ListOptions{All: true, Size: true})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(containers, func(a, b types.Container) int {
		if a.Created != b.Created {
			return cmp.Compare(a.Created, b.Created) // Descending order
		}
		return cmp.Compare(a.ID, b.ID)
	})

	eContainers := make([]extendedContainer, len(containers))

	for i, c := range containers {
		slices.SortFunc(c.Mounts, func(a, b types.MountPoint) int {
			return cmp.Compare(a.Destination, b.Destination)
		})
		slices.SortFunc(c.Ports, func(a, b types.Port) int {
			return cmp.Compare(a.PrivatePort, b.PrivatePort)
		})

		eContainers[i] = extendedContainer{Container: c}
		if len(c.Names) > 0 {
			eContainers[i].Name = strings.TrimPrefix(c.Names[0], "/")
		}
	}
	return eContainers, nil
}

func sortVolumes(ctx context.Context, dCli neTypes.DockerMetricsClient) (interface{}, error) {
	volumeSum, err := dCli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}

	volumes := volumeSum.Volumes
	slices.SortFunc(volumes, func(a, b *volume.Volume) int {
		// Convert string to timeUnix
		aCreatedAt, errA := time.Parse(time.RFC3339, a.CreatedAt)
		bCreatedAt, errB := time.Parse(time.RFC3339, b.CreatedAt)
		if errA != nil || errB != nil {
			return cmp.Compare(a.Name, b.Name)
		}

		if aCreatedAt.Unix() != bCreatedAt.Unix() {
			return cmp.Compare(aCreatedAt.Unix(), bCreatedAt.Unix()) // Descending order
		}
		return cmp.Compare(a.Name, b.Name)
	})

	return volumes, nil
}

func sortNetworks(ctx context.Context, dCli neTypes.DockerMetricsClient) (interface{}, error) {
	networks, err := dCli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(networks, func(a, b network.Summary) int {
		if a.Created != b.Created {
			return cmp.Compare(b.Created.Unix(), a.Created.Unix()) // Descending order
		}

		return cmp.Compare(a.ID, b.ID)
	})

	return networks, nil
}

func sortServices(ctx context.Context, dCli neTypes.DockerMetricsClient) (interface{}, error) {
	services, err := dCli.ServiceList(ctx, types.ServiceListOptions{Status: true})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(services, func(a, b swarm.Service) int {
		if a.CreatedAt.Unix() != b.CreatedAt.Unix() {
			return cmp.Compare(b.CreatedAt.Unix(), a.CreatedAt.Unix()) // Descending order
		}
		return cmp.Compare(a.ID, b.ID)
	})

	return services, nil
}

func sortTasks(ctx context.Context, dCli neTypes.DockerMetricsClient) (interface{}, error) {
	tasks, err := dCli.TaskList(ctx, types.TaskListOptions{})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(tasks, func(a, b swarm.Task) int {
		if a.CreatedAt.Unix() != b.CreatedAt.Unix() {
			return cmp.Compare(b.CreatedAt.Unix(), a.CreatedAt.Unix()) // Descending order
		}
		return cmp.Compare(a.ID, b.ID)
	})

	return tasks, nil
}

func sortConfigs(ctx context.Context, dCli neTypes.DockerMetricsClient) (interface{}, error) {
	configs, err := dCli.ConfigList(ctx, types.ConfigListOptions{})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(configs, func(a, b swarm.Config) int {
		if a.CreatedAt.Unix() != b.CreatedAt.Unix() {
			return cmp.Compare(b.CreatedAt.Unix(), a.CreatedAt.Unix()) // Descending order
		}
		return cmp.Compare(a.ID, b.ID)
	})

	return configs, nil
}

func sortSecrets(ctx context.Context, dCli neTypes.DockerMetricsClient) (interface{}, error) {
	secrets, err := dCli.SecretList(ctx, types.SecretListOptions{})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(secrets, func(a, b swarm.Secret) int {
		if a.CreatedAt.Unix() != b.CreatedAt.Unix() {
			return cmp.Compare(b.CreatedAt.Unix(), a.CreatedAt.Unix()) // Descending order
		}
		return cmp.Compare(a.ID, b.ID)
	})

	return secrets, nil
}

func (dm *DockerMonitor) getGatherers() []gatherer {
	gatherers := []gatherer{
		{
			false,
			"images",
			sortImages,
			&dm.coeResources.DockerResources.Images,
		},
		{
			false,
			"containers",
			sortContainers,
			&dm.coeResources.DockerResources.Containers,
		},
		{
			false,
			"volumes",
			sortVolumes,
			&dm.coeResources.DockerResources.Volumes,
		},
		{
			false,
			"networks",
			sortNetworks,
			&dm.coeResources.DockerResources.Networks,
		},
	}

	_, err := dm.client.SwarmInspect(context.Background())
	if err != nil {
		log.Warn("Swarm not found, skipping swarm resources: ", err)
		return gatherers
	}

	swarmGatherers := []gatherer{
		{
			true,
			"services",
			sortServices,
			&dm.coeResources.DockerResources.Services,
		},
		{
			true,
			"tasks",
			sortTasks,
			&dm.coeResources.DockerResources.Tasks},
		{
			true,
			"configs",
			sortConfigs,
			&dm.coeResources.DockerResources.Configs},
		{
			true,
			"secrets",
			sortSecrets,
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
		go func(g gatherer) {
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

func (dm *DockerMonitor) retrieveResources(ctx context.Context, retrieveFunc gathererFunc) ([]map[string]interface{}, error) {
	resources, err := retrieveFunc(ctx, dm.client)
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
