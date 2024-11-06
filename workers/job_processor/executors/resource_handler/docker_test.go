package resource_handler

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"nuvlaedge-go/testutils/mocks"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Disable logging
	log.SetOutput(io.Discard)

	// Run tests
	m.Run()
}

func Test_NotImplementedActionResponse_ReturnsCorrectMessage(t *testing.T) {
	t.Parallel()
	action := "testAction"
	response := NewNotImplementedActionResponse(action)
	expectedMessage := fmt.Sprintf("Action %s not implemented", action)
	assert.Equal(t, expectedMessage, response.Message)
	assert.False(t, response.Success)
	assert.Equal(t, 501, response.ReturnCode)
}

func Test_NotImplementedActionResponse_EmptyAction(t *testing.T) {
	t.Parallel()
	action := ""
	response := NewNotImplementedActionResponse(action)
	expectedMessage := "Action  not implemented"
	assert.Equal(t, expectedMessage, response.Message)
	assert.False(t, response.Success)
	assert.Equal(t, 501, response.ReturnCode)
}

func Test_NotImplementedActionResponse_NilResponse(t *testing.T) {
	t.Parallel()
	action := "testAction"
	response := NewNotImplementedActionResponse(action)
	assert.NotNil(t, response)
}

func Test_ResourceNotAvailableForAction_ReturnsCorrectMessage(t *testing.T) {
	t.Parallel()
	resource := "testResource"
	action := "testAction"
	response := NewResourceNotAvailableForAction(resource, action)
	expectedMessage := fmt.Sprintf("Resource %s not available for action %s", resource, action)
	assert.Equal(t, expectedMessage, response.Message)
	assert.False(t, response.Success)
	assert.Equal(t, 404, response.ReturnCode)
}

func Test_ResourceNotAvailableForAction_EmptyResource(t *testing.T) {
	t.Parallel()
	resource := ""
	action := "testAction"
	response := NewResourceNotAvailableForAction(resource, action)
	expectedMessage := "Resource  not available for action testAction"
	assert.Equal(t, expectedMessage, response.Message)
	assert.False(t, response.Success)
	assert.Equal(t, 404, response.ReturnCode)
}

func Test_ResourceNotAvailableForAction_EmptyAction(t *testing.T) {
	t.Parallel()
	resource := "testResource"
	action := ""
	response := NewResourceNotAvailableForAction(resource, action)
	expectedMessage := "Resource testResource not available for action "
	assert.Equal(t, expectedMessage, response.Message)
	assert.False(t, response.Success)
	assert.Equal(t, 404, response.ReturnCode)
}

func Test_ResourceNotAvailableForAction_NilResponse(t *testing.T) {
	t.Parallel()
	resource := "testResource"
	action := "testAction"
	response := NewResourceNotAvailableForAction(resource, action)
	assert.NotNil(t, response)
}

func Test_ErrorResourceActionResponse_ReturnsCorrectMessage(t *testing.T) {
	t.Parallel()
	resource := "testResource"
	action := "testAction"
	returnCode := 500
	err := fmt.Errorf("test error")
	response := NewErrorResourceActionResponse(resource, action, "", returnCode, err)
	assert.False(t, response.Success)
	assert.Equal(t, returnCode, response.ReturnCode)
}

func Test_ErrorResourceActionResponse_EmptyResource(t *testing.T) {
	t.Parallel()
	resource := ""
	action := "testAction"
	returnCode := 500
	err := fmt.Errorf("test error")
	response := NewErrorResourceActionResponse(resource, action, "", returnCode, err)
	assert.False(t, response.Success)
	assert.Equal(t, returnCode, response.ReturnCode)
}

func Test_ErrorResourceActionResponse_EmptyAction(t *testing.T) {
	t.Parallel()
	resource := "testResource"
	action := ""
	returnCode := 500
	err := fmt.Errorf("test error")
	response := NewErrorResourceActionResponse(resource, action, "", returnCode, err)
	assert.False(t, response.Success)
	assert.Equal(t, returnCode, response.ReturnCode)
}

func Test_ErrorResourceActionResponse_NilError(t *testing.T) {
	t.Parallel()
	resource := "testResource"
	action := "testAction"
	returnCode := 500
	var err error = nil
	response := NewErrorResourceActionResponse(resource, action, "", returnCode, err)
	assert.False(t, response.Success)
	assert.Equal(t, returnCode, response.ReturnCode)
}

func Test_ErrorResourceActionResponse_NilResponse(t *testing.T) {
	t.Parallel()
	resource := "testResource"
	action := "testAction"
	returnCode := 500
	err := fmt.Errorf("test error")
	response := NewErrorResourceActionResponse(resource, action, "", returnCode, err)
	assert.NotNil(t, response)
}

func Test_NewDockerResourcesHandler_ReturnsHandler(t *testing.T) {
	handler, err := NewDockerResourceHandler(nil)
	assert.NoError(t, err)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.client)
	assert.NotNil(t, handler.gathererFuncs)
}

func Test_NewDockerResourcesHandler_WithProvidedClient(t *testing.T) {
	mockClient := mocks.NewResourceHandlerDockerClient(t)
	handler, err := NewDockerResourceHandler(mockClient)
	assert.NoError(t, err)
	assert.NotNil(t, handler)
	assert.Equal(t, mockClient, handler.client)
}

func MockActionFunction(ctx context.Context, id string) (ResourceActionResponse, error) {
	return ResourceActionResponse{Success: true, ReturnCode: 200, Message: "Mock Action Function"}, nil
}

func Test_DockerResourceHandler_HandleActions(t *testing.T) {
	useCases := []struct {
		name            string
		actions         []ResourceAction
		gatherers       map[string]map[string]ResourceActionFunc
		expectedResults []ResourceActionResponse
	}{
		{
			name: "SingleValidAction",
			actions: []ResourceAction{
				{Action: "pull", Resource: "image", Id: "test-image"},
			},
			gatherers: map[string]map[string]ResourceActionFunc{
				"pull": {"image": MockActionFunction},
			},
			expectedResults: []ResourceActionResponse{
				{Success: true, ReturnCode: 200, Message: "Mock Action Function"},
			},
		},
		{
			name: "MultipleValidActions",
			actions: []ResourceAction{
				{Action: "pull", Resource: "image", Id: "test-image"},
				{Action: "remove", Resource: "container", Id: "test-container"},
			},
			gatherers: map[string]map[string]ResourceActionFunc{
				"pull":   {"image": MockActionFunction},
				"remove": {"container": MockActionFunction},
			},
			expectedResults: []ResourceActionResponse{
				{Success: true, ReturnCode: 200, Message: "Mock Action Function"},
				{Success: true, ReturnCode: 200, Message: "Mock Action Function"},
			},
		},
		{
			name: "InvalidAction",
			actions: []ResourceAction{
				{Action: "invalid", Resource: "image", Id: "test-image"},
			},
			gatherers: map[string]map[string]ResourceActionFunc{},
			expectedResults: []ResourceActionResponse{
				{Success: false, ReturnCode: 501, Message: "Action invalid not implemented"},
			},
		},
		{
			name: "InvalidResource",
			actions: []ResourceAction{
				{Action: "pull", Resource: "invalid", Id: "test-image"},
			},
			gatherers: map[string]map[string]ResourceActionFunc{
				"pull": {},
			},
			expectedResults: []ResourceActionResponse{
				{Success: false, ReturnCode: 404, Message: "Resource invalid not available for action pull"},
			},
		},
		{
			name: "ErrorDuringAction",
			actions: []ResourceAction{
				{Action: "remove", Resource: "container", Id: "non-existent-container"},
			},
			gatherers: map[string]map[string]ResourceActionFunc{
				"remove": {"container": func(ctx context.Context, id string) (ResourceActionResponse, error) {
					return ResourceActionResponse{}, fmt.Errorf("test error")
				}},
			},
			expectedResults: []ResourceActionResponse{
				{Success: false, ReturnCode: 500, Message: "Error performing action remove on resource container: test error"},
			},
		},
	}

	for _, uc := range useCases {
		t.Run(uc.name, func(t *testing.T) {
			mockClient := mocks.NewResourceHandlerDockerClient(t)
			handler, err := NewDockerResourceHandler(mockClient)
			assert.NoError(t, err)
			assert.NotNil(t, handler)

			handler.gathererFuncs = uc.gatherers

			responses := handler.HandleActions(context.Background(), uc.actions)
			assert.Equal(t, len(uc.expectedResults), len(responses))
			for i, expected := range uc.expectedResults {
				assert.Equal(t, expected.Success, responses[i].Success)
				assert.Equal(t, expected.ReturnCode, responses[i].ReturnCode)
			}
		})
	}
}

func Test_DockerResourceHandler_HandleAction_ValidAction(t *testing.T) {
	mockClient := mocks.NewResourceHandlerDockerClient(t)
	handler, err := NewDockerResourceHandler(mockClient)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	mockGatherers := map[string]map[string]ResourceActionFunc{
		"pull": {
			"image": MockActionFunction,
		},
	}
	handler.gathererFuncs = mockGatherers

	action := ResourceAction{Action: "pull", Resource: "image", Id: "test-image"}
	response := handler.handleAction(context.Background(), action)
	assert.True(t, response.Success)
	assert.Equal(t, 200, response.ReturnCode)
	assert.Equal(t, "Mock Action Function", response.Message)
}

func Test_DockerResourceHandler_HandleAction_InvalidAction(t *testing.T) {
	mockClient := mocks.NewResourceHandlerDockerClient(t)
	handler, err := NewDockerResourceHandler(mockClient)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	mockGatherers := map[string]map[string]ResourceActionFunc{}
	handler.gathererFuncs = mockGatherers

	action := ResourceAction{Action: "invalid", Resource: "image", Id: "test-image"}
	response := handler.handleAction(context.Background(), action)
	assert.False(t, response.Success)
	assert.Equal(t, 501, response.ReturnCode)
	assert.Equal(t, "Action invalid not implemented", response.Message)
}

func Test_DockerResourceHandler_HandleAction_InvalidResource(t *testing.T) {
	mockClient := mocks.NewResourceHandlerDockerClient(t)
	handler, err := NewDockerResourceHandler(mockClient)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	mockGatherers := map[string]map[string]ResourceActionFunc{
		"pull": {},
	}
	handler.gathererFuncs = mockGatherers

	action := ResourceAction{Action: "pull", Resource: "invalid", Id: "test-image"}
	response := handler.handleAction(context.Background(), action)
	assert.False(t, response.Success)
	assert.Equal(t, 404, response.ReturnCode)
	assert.Equal(t, "Resource invalid not available for action pull", response.Message)
}

func Test_DockerResourceHandler_HandleAction(t *testing.T) {
	useCases := []struct {
		name            string
		action          ResourceAction
		gatherers       map[string]map[string]ResourceActionFunc
		success         bool
		returnCode      int
		messageContains string
	}{
		{
			name:   "ValidAction",
			action: ResourceAction{Action: "pull", Resource: "image", Id: "test-image"},
			gatherers: map[string]map[string]ResourceActionFunc{
				"pull": {"image": MockActionFunction},
			},
			success:         true,
			returnCode:      200,
			messageContains: "Mock Action Function",
		},
		{
			name:            "InvalidAction",
			action:          ResourceAction{Action: "invalid", Resource: "image", Id: "test-image"},
			gatherers:       map[string]map[string]ResourceActionFunc{},
			success:         false,
			returnCode:      501,
			messageContains: "Action invalid not implemented",
		},
		{
			name:   "InvalidResource",
			action: ResourceAction{Action: "pull", Resource: "invalid", Id: "test-image"},
			gatherers: map[string]map[string]ResourceActionFunc{
				"pull": {},
			},
			success:         false,
			returnCode:      404,
			messageContains: "Resource invalid not available for action pull",
		},
		{
			name:   "ErrorDuringAction",
			action: ResourceAction{Action: "remove", Resource: "container", Id: "non-existent-container"},
			gatherers: map[string]map[string]ResourceActionFunc{
				"remove": {"container": func(ctx context.Context, id string) (ResourceActionResponse, error) {
					return ResourceActionResponse{}, fmt.Errorf("test error")
				}},
			},
			success:         false,
			returnCode:      500,
			messageContains: "Error performing action remove on resource container: test error",
		},
		{
			name:            "EmptyAction",
			action:          ResourceAction{Action: "", Resource: "image", Id: "test-image"},
			gatherers:       map[string]map[string]ResourceActionFunc{},
			success:         false,
			returnCode:      501,
			messageContains: "Action  not implemented",
		},
	}

	for _, uc := range useCases {
		t.Run(uc.name, func(t *testing.T) {
			mockClient := mocks.NewResourceHandlerDockerClient(t)
			handler, err := NewDockerResourceHandler(mockClient)
			assert.NoError(t, err)
			assert.NotNil(t, handler)

			handler.gathererFuncs = uc.gatherers

			response := handler.handleAction(context.Background(), uc.action)
			assert.Equal(t, uc.success, response.Success)
			assert.Equal(t, uc.returnCode, response.ReturnCode)
		})
	}
}

func mockReadCloser(content string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(content))
}

func Test_DockerResourceHandler_PullImage(t *testing.T) {
	useCases := []struct {
		name            string
		imageID         string
		mockReturn      io.ReadCloser
		mockError       error
		success         bool
		returnCode      int
		messageContains string
	}{
		{
			name:            "SuccessfulDownload",
			imageID:         "test-image",
			mockReturn:      mockReadCloser("Downloaded newer image"),
			mockError:       nil,
			success:         true,
			returnCode:      200,
			messageContains: "Image test-image downloaded successfully",
		},
		{
			name:            "UpToDate",
			imageID:         "test-image",
			mockReturn:      mockReadCloser("Image is up to date"),
			mockError:       nil,
			success:         true,
			returnCode:      200,
			messageContains: "Image test-image is up to date",
		},
		{
			name:            "ErrorDuringPull",
			imageID:         "test-image",
			mockReturn:      nil,
			mockError:       fmt.Errorf("test error"),
			success:         false,
			returnCode:      500,
			messageContains: "Error performing action pull on resource image: test error",
		},
	}

	for _, uc := range useCases {
		t.Run(uc.name, func(t *testing.T) {
			mockClient := mocks.NewResourceHandlerDockerClient(t)
			handler, err := NewDockerResourceHandler(mockClient)
			assert.NoError(t, err)
			assert.NotNil(t, handler)

			mockClient.On("ImagePull", mock.Anything, uc.imageID, mock.Anything).Return(uc.mockReturn, uc.mockError)

			response, err := handler.pullImage(context.Background(), uc.imageID)
			if uc.mockError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if uc.mockReturn != nil {
				assert.Equal(t, uc.success, response.Success)
				assert.Equal(t, uc.returnCode, response.ReturnCode)
				assert.Contains(t, response.Message, uc.messageContains)
			}
		})
	}
}

func Test_DockerResourceHandler_RemoveImage(t *testing.T) {
	useCases := []struct {
		name            string
		imageID         string
		mockReturn      []image.DeleteResponse
		mockError       error
		success         bool
		returnCode      int
		messageContains string
	}{
		{
			name:            "SuccessfulDeletion",
			imageID:         "test-image",
			mockReturn:      []image.DeleteResponse{{Deleted: "test-image"}},
			mockError:       nil,
			success:         true,
			returnCode:      200,
			messageContains: "Image test-image deleted successfully",
		},
		{
			name:            "SuccessfulUntag",
			imageID:         "test-image",
			mockReturn:      []image.DeleteResponse{{Untagged: "test-image"}},
			mockError:       nil,
			success:         true,
			returnCode:      200,
			messageContains: "Image test-image untagged but not removed",
		},
		{
			name:            "ErrorDuringRemoval",
			imageID:         "test-image",
			mockReturn:      nil,
			mockError:       fmt.Errorf("test error"),
			success:         false,
			returnCode:      500,
			messageContains: "Error performing action remove on resource image: test error",
		},
	}

	for _, uc := range useCases {
		t.Run(uc.name, func(t *testing.T) {
			mockClient := mocks.NewResourceHandlerDockerClient(t)
			handler, err := NewDockerResourceHandler(mockClient)
			assert.NoError(t, err)
			assert.NotNil(t, handler)

			mockClient.On("ImageRemove", mock.Anything, uc.imageID, mock.Anything).Return(uc.mockReturn, uc.mockError)

			response, err := handler.removeImage(context.Background(), uc.imageID)
			if uc.mockError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if err == nil {
				assert.Equal(t, uc.success, response.Success)
				assert.Equal(t, uc.returnCode, response.ReturnCode)
				assert.Contains(t, response.Message, uc.messageContains)
			}
		})
	}
}

func Test_DockerResourceHandler_RemoveContainer(t *testing.T) {
	useCases := []struct {
		name            string
		containerID     string
		mockError       error
		success         bool
		returnCode      int
		messageContains string
	}{
		{
			name:            "SuccessfulRemoval",
			containerID:     "test-container",
			mockError:       nil,
			success:         true,
			returnCode:      204,
			messageContains: "Container test-container removed successfully",
		},
		{
			name:            "ErrorDuringRemoval",
			containerID:     "test-container",
			mockError:       fmt.Errorf("test error"),
			success:         false,
			returnCode:      500,
			messageContains: "Error performing action remove on resource container: test error",
		},
	}

	for _, uc := range useCases {
		t.Run(uc.name, func(t *testing.T) {
			mockClient := mocks.NewResourceHandlerDockerClient(t)
			handler, err := NewDockerResourceHandler(mockClient)
			assert.NoError(t, err)
			assert.NotNil(t, handler)

			mockClient.On("ContainerRemove", mock.Anything, uc.containerID, mock.Anything).Return(uc.mockError)

			response, err := handler.removeContainer(context.Background(), uc.containerID)
			if uc.mockError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if err == nil {
				assert.Equal(t, uc.success, response.Success)
				assert.Equal(t, uc.returnCode, response.ReturnCode)
				assert.Contains(t, response.Message, uc.messageContains)
			}
		})
	}
}

func Test_DockerResourceHandler_RemoveVolume(t *testing.T) {
	useCases := []struct {
		name            string
		volumeID        string
		mockError       error
		success         bool
		returnCode      int
		messageContains string
	}{
		{
			name:            "SuccessfulRemoval",
			volumeID:        "test-volume",
			mockError:       nil,
			success:         true,
			returnCode:      204,
			messageContains: "Volume test-volume removed successfully",
		},
		{
			name:            "ErrorDuringRemoval",
			volumeID:        "test-volume",
			mockError:       fmt.Errorf("test error"),
			success:         false,
			returnCode:      500,
			messageContains: "Error performing action remove on resource volume: test error",
		},
	}

	for _, uc := range useCases {
		t.Run(uc.name, func(t *testing.T) {
			mockClient := mocks.NewResourceHandlerDockerClient(t)
			handler, err := NewDockerResourceHandler(mockClient)
			assert.NoError(t, err)
			assert.NotNil(t, handler)

			mockClient.On("VolumeRemove", mock.Anything, uc.volumeID, false).Return(uc.mockError)

			response, err := handler.removeVolume(context.Background(), uc.volumeID)
			if uc.mockError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if err == nil {
				assert.Equal(t, uc.success, response.Success)
				assert.Equal(t, uc.returnCode, response.ReturnCode)
				assert.Contains(t, response.Message, uc.messageContains)
			}
		})
	}
}

func Test_DockerResourceHandler_RemoveNetwork(t *testing.T) {
	useCases := []struct {
		name            string
		networkID       string
		mockError       error
		success         bool
		returnCode      int
		messageContains string
	}{
		{
			name:            "SuccessfulRemoval",
			networkID:       "test-network",
			mockError:       nil,
			success:         true,
			returnCode:      204,
			messageContains: "Network test-network removed successfully",
		},
		{
			name:            "ErrorDuringRemoval",
			networkID:       "test-network",
			mockError:       fmt.Errorf("test error"),
			success:         false,
			returnCode:      500,
			messageContains: "Error performing action remove on resource network: test error",
		},
	}

	for _, uc := range useCases {
		t.Run(uc.name, func(t *testing.T) {
			mockClient := mocks.NewResourceHandlerDockerClient(t)
			handler, err := NewDockerResourceHandler(mockClient)
			assert.NoError(t, err)
			assert.NotNil(t, handler)

			mockClient.On("NetworkRemove", mock.Anything, uc.networkID).Return(uc.mockError)

			response, err := handler.removeNetwork(context.Background(), uc.networkID)
			if uc.mockError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if err == nil {
				assert.Equal(t, uc.success, response.Success)
				assert.Equal(t, uc.returnCode, response.ReturnCode)
				assert.Contains(t, response.Message, uc.messageContains)
			}
		})
	}
}

func Test_GetCodeFromError(t *testing.T) {
	useCases := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "Returns400ForInvalidParameterError",
			err:      errdefs.InvalidParameter(fmt.Errorf("invalid parameter")),
			expected: 400,
		},
		{
			name:     "Returns404ForNotFoundError",
			err:      errdefs.NotFound(fmt.Errorf("not found")),
			expected: 404,
		},
		{
			name:     "Returns409ForConflictError",
			err:      errdefs.Conflict(fmt.Errorf("conflict")),
			expected: 409,
		},
		{
			name:     "Returns500ForUnknownError",
			err:      fmt.Errorf("unknown error"),
			expected: 500,
		},
	}

	for _, uc := range useCases {
		t.Run(uc.name, func(t *testing.T) {
			code := getCodeFromError(uc.err)
			assert.Equal(t, uc.expected, code)
		})
	}
}
