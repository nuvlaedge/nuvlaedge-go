package resource_handler

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	log "github.com/sirupsen/logrus"
	"strings"
)

type ResourceAction struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
	Id       string `json:"id"`
}

type ResourceActionResponse struct {
	Success    bool        `json:"success"`
	ReturnCode int         `json:"return-code"`
	Message    string      `json:"message,omitempty"`
	Content    interface{} `json:"content,omitempty"`
}

func NewNotImplementedActionResponse(action string) *ResourceActionResponse {
	msg := fmt.Sprintf("Action %s not implemented", action)
	return NewResourceActionResponse(false, 501, msg)
}

func NewResourceNotAvailableForAction(resource, action string) *ResourceActionResponse {
	msg := fmt.Sprintf("Resource %s not available for action %s", resource, action)
	return NewResourceActionResponse(false, 404, msg)
}

func NewErrorResourceActionResponse(resource, action string, returnCode int, err error) *ResourceActionResponse {
	msg := fmt.Sprintf("Error performing action %s on resource %s: %s", action, resource, err)
	return NewResourceActionResponse(false, returnCode, msg)
}

func NewResourceActionResponse(success bool, returnCode int, message string) *ResourceActionResponse {
	return &ResourceActionResponse{
		Success:    success,
		ReturnCode: returnCode,
		Message:    message,
	}
}

type ResourceActionFunc func(ctx context.Context, id string) (ResourceActionResponse, error)

type DockerResourceHandler struct {
	client ResourceHandlerDockerClient

	gathererFuncs map[string]map[string]ResourceActionFunc
}

func NewDockerResourceHandler(dCli ResourceHandlerDockerClient) (*DockerResourceHandler, error) {
	var err error
	if dCli == nil {
		dCli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, err
		}
	}

	d := &DockerResourceHandler{
		client: dCli,
	}

	d.gathererFuncs = map[string]map[string]ResourceActionFunc{
		"pull": {
			"image": d.pullImage,
		},
		"remove": {
			"image":     d.removeImage,
			"container": d.removeContainer,
			"volume":    d.removeVolume,
			"network":   d.removeNetwork,
		},
	}

	return d, nil
}

func (drh *DockerResourceHandler) HandleActions(ctx context.Context, actions []ResourceAction) []ResourceActionResponse {
	responses := make([]ResourceActionResponse, len(actions))

	for i, action := range actions {
		log.Infof("Handling action %s on resource %s with id %s", action.Action, action.Resource, action.Id)
		responses[i] = drh.handleAction(ctx, action)
	}

	return responses
}

func (drh *DockerResourceHandler) handleAction(ctx context.Context, action ResourceAction) ResourceActionResponse {
	actionFunc, response := drh.getActionFunc(action)
	if response != nil {
		return *response
	}

	resp, err := actionFunc(ctx, action.Id)
	if err != nil {
		return *NewErrorResourceActionResponse(action.Resource, action.Action, getCodeFromError(err), err)
	}

	return resp
}

func (drh *DockerResourceHandler) getActionFunc(action ResourceAction) (ResourceActionFunc, *ResourceActionResponse) {
	gatherer, ok := drh.gathererFuncs[action.Action]
	if !ok {
		return nil, NewNotImplementedActionResponse(action.Action)
	}

	actionFunc, ok := gatherer[action.Resource]
	if !ok {
		return nil, NewResourceNotAvailableForAction(action.Resource, action.Action)
	}

	return actionFunc, nil
}

func (drh *DockerResourceHandler) pullImage(ctx context.Context, id string) (ResourceActionResponse, error) {
	rCloser, err := drh.client.ImagePull(ctx, id, image.PullOptions{})
	if err != nil {
		return ResourceActionResponse{}, err
	}
	defer rCloser.Close()

	// Read the response
	var lastLine string
	scanner := bufio.NewScanner(rCloser)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	if strings.Contains(lastLine, "Downloaded newer image") {
		msg := fmt.Sprintf("Image %s downloaded successfully", id)
		return *NewResourceActionResponse(true, 200, msg), nil
	}

	if strings.Contains(lastLine, "Image is up to date") {
		msg := fmt.Sprintf("Image %s is up to date", id)
		return *NewResourceActionResponse(true, 200, msg), nil
	}

	return *NewResourceActionResponse(true, 200, "Image pull successful"), nil
}

func (drh *DockerResourceHandler) removeImage(ctx context.Context, id string) (ResourceActionResponse, error) {
	res, err := drh.client.ImageRemove(ctx, id, image.RemoveOptions{})
	if err != nil {
		return ResourceActionResponse{}, err
	}
	var deleted, untagged bool
	for _, r := range res {
		if strings.Contains(r.Deleted, id) {
			fmt.Printf("Deleted: %s\n", r.Deleted)
			deleted = true
		}
		if strings.Contains(r.Untagged, id) {
			untagged = true
		}
	}

	if deleted {
		msg := fmt.Sprintf("Image %s deleted successfully", id)
		return *NewResourceActionResponse(true, 200, msg), nil
	}

	if untagged {
		msg := fmt.Sprintf("Image %s untagged but not removed", id)
		return *NewResourceActionResponse(true, 200, msg), nil
	}

	return *NewResourceActionResponse(true, 200, ""), nil
}

func (drh *DockerResourceHandler) removeContainer(ctx context.Context, id string) (ResourceActionResponse, error) {
	err := drh.client.ContainerRemove(ctx, id, container.RemoveOptions{})
	if err != nil {
		return ResourceActionResponse{}, err
	}

	msg := fmt.Sprintf("Container %s removed successfully", id)

	return *NewResourceActionResponse(true, 204, msg), nil
}

func (drh *DockerResourceHandler) removeVolume(ctx context.Context, id string) (ResourceActionResponse, error) {
	err := drh.client.VolumeRemove(ctx, id, false)
	if err != nil {
		return ResourceActionResponse{}, err
	}

	msg := fmt.Sprintf("Volume %s removed successfully", id)

	return *NewResourceActionResponse(true, 204, msg), nil
}

func (drh *DockerResourceHandler) removeNetwork(ctx context.Context, id string) (ResourceActionResponse, error) {
	err := drh.client.NetworkRemove(ctx, id)
	if err != nil {
		return ResourceActionResponse{}, err
	}

	msg := fmt.Sprintf("Network %s removed successfully", id)

	return *NewResourceActionResponse(true, 204, msg), nil
}

func getCodeFromError(err error) int {
	switch err.(type) {
	case errdefs.ErrInvalidParameter:
		return 400
	case errdefs.ErrNotFound:
		return 404
	case errdefs.ErrConflict:
		return 409

	default:
		log.Warnf("Unknown error type: %T", err)
		return 500
	}
}
