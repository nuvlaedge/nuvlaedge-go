package workers

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types/filters"
	log "github.com/sirupsen/logrus"
	"nuvlaedge-go/types"
	"nuvlaedge-go/types/worker"
	"slices"
)

type ResourceCleaner interface {
	worker.Worker // ResourceCleaner is a worker

	//
}

type DockerCleaner struct {
	worker.TimedWorker
	dCli types.CleanerClient

	// Objects to clean
	// Options are:
	// - containers
	// - images
	// - volumes
	// - networks
	// - system
	objects []string

	clearnerFactory map[string]func(ctx context.Context) error
}

func (d *DockerCleaner) Init(opts *worker.WorkerOpts, conf *worker.WorkerConfig) error {
	d.TimedWorker = worker.NewTimedWorker(conf.CleanUpPeriod, worker.ResourceCleaner)
	d.dCli = opts.DockerClient
	d.objects = conf.RemoveObjects
	return nil
}

func (d *DockerCleaner) Start(ctx context.Context) error {
	d.clearnerFactory = map[string]func(ctx context.Context) error{
		"containers": d.cleanContainers,
		"images":     d.cleanImages,
		"volumes":    d.cleanVolumes,
		"networks":   d.cleanNetworks,
		"system":     d.cleanSystem,
	}

	go func() {
		err := d.Run(ctx)
		if err != nil {
			log.Errorf("Error running Commissioner: %s", err)
		}
	}()
	return nil
}

func (d *DockerCleaner) Reconfigure(conf *worker.WorkerConfig) error {
	// Do this check to prevent the ticker from being reset
	if conf.CleanUpPeriod != d.GetPeriod() {
		d.SetPeriod(conf.CleanUpPeriod)
	}
	d.objects = conf.RemoveObjects
	return nil
}

func (d *DockerCleaner) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			if err := d.Stop(ctx); err != nil {
				return err
			}
			return ctx.Err()

		case <-d.BaseTicker.C:
			if err := d.cleanResources(ctx); err != nil {
				log.Error("Failed to clean resources: ", err)
			}
		case conf := <-d.ConfChan:
			log.Debug("Received configuration in cleaner: ", conf)
			if err := d.Reconfigure(conf); err != nil {
				log.Error("Failed to reconfigure DockerCleaner: ", err)
			}
		}
	}
}

func (d *DockerCleaner) Stop(_ context.Context) error {
	d.BaseTicker.Stop()
	return nil
}

func (d *DockerCleaner) cleanResources(ctx context.Context) error {
	if len(d.objects) == 0 {
		// Prune dangling images only
		return d.cleanImages(ctx)
	}

	var errList []error
	if slices.Contains(d.objects, "system") {
		// If system is in the list, clean all resources and ignore the rest unless volumes are
		if slices.Contains(d.objects, "volumes") {
			if err := d.cleanVolumes(ctx); err != nil {
				errList = append(errList, err)
			}

			if err := d.cleanSystem(ctx); err != nil {
				errList = append(errList, err)
			}

			return errors.Join(errList...)
		}
	}

	for _, obj := range d.objects {
		log.Infof("Cleaning %s", obj)

		f, ok := d.clearnerFactory[obj]
		if !ok {
			log.Warnf("No cleaner for %s", obj)
			continue
		}

		if err := f(ctx); err != nil {
			errList = append(errList, err)
		}
	}

	return errors.Join(errList...)
}

func (d *DockerCleaner) cleanContainers(ctx context.Context) error {
	// Remove all stopped containers
	rep, err := d.dCli.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		return err
	}
	log.Infof("Removed %v stopped containers", rep.ContainersDeleted)
	log.Infof("Reclaimed %d bytes", rep.SpaceReclaimed)
	return nil
}

func (d *DockerCleaner) cleanImages(ctx context.Context) error {
	rep, err := d.dCli.ImagesPrune(ctx, filters.Args{})
	if err != nil {
		return err
	}
	log.Infof("Removed %v images", rep.ImagesDeleted)
	log.Infof("Reclaimed %d bytes", rep.SpaceReclaimed)
	return nil
}

func (d *DockerCleaner) cleanVolumes(ctx context.Context) error {
	rep, err := d.dCli.VolumesPrune(ctx, filters.Args{})
	if err != nil {
		return err
	}

	log.Infof("Removed %v volumes", rep.VolumesDeleted)
	log.Infof("Reclaimed %d bytes", rep.SpaceReclaimed)

	return nil
}

func (d *DockerCleaner) cleanNetworks(ctx context.Context) error {
	rep, err := d.dCli.NetworksPrune(ctx, filters.Args{})
	if err != nil {
		return err
	}
	log.Infof("Removed %v networks", rep.NetworksDeleted)
	return nil
}

func (d *DockerCleaner) cleanSystem(ctx context.Context) error {
	var errList []error
	if err := d.cleanContainers(ctx); err != nil {
		errList = append(errList, err)
	}
	if err := d.cleanImages(ctx); err != nil {
		errList = append(errList, err)
	}
	if err := d.cleanNetworks(ctx); err != nil {
		errList = append(errList, err)
	}
	return errors.Join(errList...)
}
