package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

type testDockerDaemon struct {
	Port               int
	Host               string
	ContainerID        string
	dockerHostClient   *client.Client
	dockerDaemonClient *client.Client
}

func startContainer(t testing.TB, cl *client.Client, image string, cmd []string, ports []string, privileged bool, labels map[string]string) container.ContainerCreateCreatedBody {
	portSet := nat.PortSet{}
	portMap := nat.PortMap{}

	reader, err := cl.ImagePull(context.Background(), image, types.ImagePullOptions{})
	assert.NoError(t, err)
	io.Copy(os.Stdout, reader)

	for _, port := range ports {
		portSet[nat.Port(port)] = struct{}{}
		portMap[nat.Port(port)] = []nat.PortBinding{{HostIP: "0.0.0.0"}}
	}

	config := &container.Config{
		AttachStdin:  true,
		Image:        image,
		Cmd:          cmd,
		ExposedPorts: portSet,
		Labels:       labels,
	}
	hostConfig := &container.HostConfig{
		PortBindings: portMap,
		Privileged:   true,
	}
	networkingConfig := &network.NetworkingConfig{}
	created, err := cl.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, "")
	assert.NoError(t, err)
	err = cl.ContainerStart(context.Background(), created.ID, types.ContainerStartOptions{})
	assert.NoError(t, err)
	go func() {
		read, err := cl.ContainerLogs(context.Background(), created.ID, types.ContainerLogsOptions{Follow: true, ShowStdout: true, ShowStderr: true})
		assert.NoError(t, err)
		if err == nil {
			io.Copy(os.Stdout, read)
		}
	}()
	return created
}

func startDockerDaemon(t testing.TB) *testDockerDaemon {
	cl, err := client.NewEnvClient()
	assert.NoError(t, err)
	info, err := cl.Info(context.Background())
	if err != nil {
		t.Skipf("Could not connect to docker host: %s", err.Error())
		t.SkipNow()
	}
	t.Logf("Launching containerized docker daemon in %s", info.Name)
	created := startContainer(t, cl, "docker:dind", []string{"dockerd", "-H", "tcp://0.0.0.0:2345"}, []string{"2345"}, true, map[string]string{})
	fmt.Println(created.ID)
	inspect, err := cl.ContainerInspect(context.Background(), created.ID)
	assert.NoError(t, err)
	port, err := strconv.Atoi(inspect.NetworkSettings.Ports["2345/tcp"][0].HostPort)
	assert.NoError(t, err)
	daemonClient, err := client.NewClient("tcp://localhost:"+inspect.NetworkSettings.Ports["2345/tcp"][0].HostPort, "", nil, map[string]string{})

	for {
		networks, err := daemonClient.NetworkList(context.Background(), types.NetworkListOptions{})
		if err == nil {
			if len(networks) > 0 {
				fmt.Println(networks)
				break
			}
		}
	}
	assert.NoError(t, err)
	return &testDockerDaemon{
		Port:               port,
		ContainerID:        created.ID,
		dockerHostClient:   cl,
		dockerDaemonClient: daemonClient,
	}
}

func (d *testDockerDaemon) stop() {
	d.dockerHostClient.ContainerKill(context.Background(), d.ContainerID, "")
	d.dockerHostClient.ContainerRemove(context.Background(), d.ContainerID, types.ContainerRemoveOptions{RemoveVolumes: true})
}

func count(l []string, k string) int {
	c := 0
	for _, v := range l {
		if v == k {
			c++
		}
	}
	return c
}

func TestNoSwarm(t *testing.T) {
	h := startDockerDaemon(t)
	defer h.stop()
	startContainer(t, h.dockerDaemonClient, "alpine", []string{"tail", "-f", "/dev/null"}, []string{}, false, map[string]string{"traefik.frontend.rule": "Host: traefik.io, www.traefik.io"})
	startContainer(t, h.dockerDaemonClient, "alpine", []string{"tail", "-f", "/dev/null"}, []string{}, false, map[string]string{"traefik.frontend.rule": "Host: www.example.com, traefik.io"})
	lister := Lister{h.dockerDaemonClient}
	domains, err := lister.List()
	assert.NoError(t, err)
	assert.Equal(t, 1, count(domains, "traefik.io"))
	assert.Equal(t, 1, count(domains, "www.traefik.io"))
	assert.Equal(t, 1, count(domains, "www.example.com"))
}

func TestSwarm(t *testing.T) {
	h := startDockerDaemon(t)
	defer h.stop()

	_, err := h.dockerDaemonClient.SwarmInit(context.Background(), swarm.InitRequest{ListenAddr: "0.0.0.0:2377"})
	assert.NoError(t, err)

	startContainer(t, h.dockerDaemonClient, "alpine", []string{"tail", "-f", "/dev/null"}, []string{}, false, map[string]string{"traefik.frontend.rule": "Host: traefik.io, www.traefik.io"})
	startContainer(t, h.dockerDaemonClient, "alpine", []string{"tail", "-f", "/dev/null"}, []string{}, false, map[string]string{"traefik.frontend.rule": "Host: www.example.com, traefik.io"})

	specs := swarm.ServiceSpec{
		Annotations: swarm.Annotations{Labels: map[string]string{"traefik.frontend.rule": "Host: swarm.example.com, traefik.io"}},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: swarm.ContainerSpec{
				Image:   "alpine",
				Command: []string{"tail", "-f", "/dev/null"},
			},
		},
	}
	r, err := h.dockerDaemonClient.ServiceCreate(context.Background(), specs, types.ServiceCreateOptions{})
	assert.NoError(t, err)
	fmt.Println(r)
	lister := Lister{h.dockerDaemonClient}
	domains, err := lister.List()
	assert.NoError(t, err)
	fmt.Println(domains)
	assert.Equal(t, 1, count(domains, "traefik.io"))
	assert.Equal(t, 1, count(domains, "www.traefik.io"))
	assert.Equal(t, 1, count(domains, "swarm.example.com"))
}
