package docker

import (
	"context"
	"github.com/joho/godotenv"
	"io"
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
	err := godotenv.Load() // TODO : move this to the main entry file
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	handleUnexpectedError(err)
	timeoutStr := os.Getenv("CONTAINER_TIMEOUT")
	timeout, err := strconv.Atoi(timeoutStr)
	return &Adapter{serviceName: serviceName, client: cli, containerRestartTimeout: timeout}, nil
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
