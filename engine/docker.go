package engine

import "github.com/docker/docker/client"

type DockerEngine struct {
	client *client.APIClient
}
