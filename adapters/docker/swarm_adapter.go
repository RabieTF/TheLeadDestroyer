package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

type Adapter struct {
	client                  *client.Client
	serviceName             string
	containerRestartTimeout int
}

func New(serviceName string) (*Adapter, error) {
	// Initialize Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Docker client: %v", err)
	}

	// Parse container timeout from environment variables
	timeoutStr := getEnvOrDefault("CONTAINER_TIMEOUT", "10")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CONTAINER_TIMEOUT value: %v", err)
	}

	// Check if the service exists
	ctx := context.Background()
	services, err := cli.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %v", err)
	}

	serviceExists := false
	for _, service := range services {
		if service.Spec.Name == serviceName {
			serviceExists = true
			break
		}
	}

	// Create the service if it doesn't exist
	if !serviceExists {
		log.Printf("Service %s not found. Creating service...\n", serviceName)

		serviceSpec := swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Name: serviceName,
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image: "servuc/hash_extractor:latest",
					Command: []string{
						"s",
						"ws://127.0.0.1:3000", // Start in slave mode with WebSocket connection
					},
				},
				RestartPolicy: &swarm.RestartPolicy{
					Condition: swarm.RestartPolicyConditionAny,
				},
			},
			Mode: swarm.ServiceMode{
				Replicated: &swarm.ReplicatedService{
					Replicas: uint64Ptr(1), // Start with 1 replica
				},
			},
		}

		_, err := cli.ServiceCreate(ctx, serviceSpec, types.ServiceCreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create service: %v", err)
		}

		log.Printf("Service %s created successfully.\n", serviceName)
	} else {
		log.Printf("Service %s already exists.\n", serviceName)
	}

	return &Adapter{
		client:                  cli,
		serviceName:             serviceName,
		containerRestartTimeout: timeout,
	}, nil
}

// Helper function to get environment variables with default values
func getEnvOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// Helper function to create a uint64 pointer
func uint64Ptr(i uint64) *uint64 {
	return &i
}

func (d *Adapter) GetServiceDetails(ctx context.Context) *swarm.Service {
	services, err := d.client.ServiceList(ctx, types.ServiceListOptions{})
	handleUnexpectedError(err)

	for _, service := range services {
		if service.Spec.Name == d.serviceName {
			return &service
		}
	}
	handleUnfoundServiceName(d.serviceName)
	return nil
}

func (d *Adapter) ScaleService(ctx context.Context, replicas uint64) error {
	service := d.GetServiceDetails(ctx)

	service.Spec.Mode.Replicated.Replicas = &replicas
	_, err := d.client.ServiceUpdate(ctx, service.ID, service.Version, service.Spec, types.ServiceUpdateOptions{})

	return err
}

func (d *Adapter) GetServiceTasks(ctx context.Context) ([]swarm.Task, error) {
	tasks, err := d.client.TaskList(ctx, types.TaskListOptions{})
	handleUnexpectedError(err)

	var serviceTasks []swarm.Task
	for _, task := range tasks {
		if task.ServiceID == d.serviceName {
			serviceTasks = append(serviceTasks, task)
		}
	}
	return serviceTasks, nil
}

func (d *Adapter) RestartTask(ctx context.Context, containerID string) error {

	err := d.client.ContainerStop(ctx, containerID, container.StopOptions{
		Signal:  "SIGTERM",
		Timeout: &d.containerRestartTimeout,
	})
	if err != nil {
		return err
	}

	return d.client.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (d *Adapter) GetContainerLogs(ctx context.Context, containerID string) (string, error) {
	logReader, err := d.client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
	})
	handleUnexpectedError(err)
	defer func(logReader io.ReadCloser) {
		err := logReader.Close()
		handleUnexpectedError(err)
	}(logReader)

	logs, err := io.ReadAll(logReader)
	handleUnexpectedError(err)
	return string(logs), nil
}

func (d *Adapter) GetContainerIPs(ctx context.Context) ([]string, error) {
	var ips []string

	// Fetch tasks for the service
	tasks, err := d.GetServiceTasks(ctx)
	handleUnexpectedError(err)

	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning {
			containerID := task.Status.ContainerStatus.ContainerID

			containerDetails, err := d.client.ContainerInspect(ctx, containerID)
			if err != nil {
				return nil, err
			}

			networkName := "ingress" // TODO: check if this will be correct upon creating
			if network, ok := containerDetails.NetworkSettings.Networks[networkName]; ok {
				ips = append(ips, network.IPAddress)
			}
		}
	}

	return ips, nil
}
