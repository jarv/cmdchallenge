package dashboard

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/sirupsen/logrus"
)

const (
	cmdNetworkName    = "cmd-network"
	captureTimeout    = 20
	cleanupTimeout    = 5
	pupeteerDockerImg = "registry.gitlab.com/jarv/cmdchallenge/puppeteer"
	pupeteerName      = "pupeteer"
	grafanaDockerImg  = "registry.gitlab.com/jarv/cmdchallenge/cmd-dashboard"
	grafanaName       = "grafana"
	dashboardURL      = "http://" + grafanaName + ":3000/d/9dMXL2N7z/cmd-application?orgId=1&from=now-24h&to=now&kiosk"
)

var (
	puppeteerCmd = []string{
		"node",
		"/app/screenshot.js",
		dashboardURL,
		"screenshot.png",
	}
)

type Dashboard struct {
	log *logrus.Logger
	cli *client.Client
}

func New(logger *logrus.Logger) *Dashboard {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &Dashboard{logger, cli}
}

func (d *Dashboard) Capture() error {
	// Create the network
	ctx, cancel := context.WithTimeout(context.Background(), captureTimeout*time.Second)
	defer cancel()

	if err := d.createNetwork(ctx); err != nil {
		return err
	}
	defer d.removeNetwork()

	// Start Grafana
	grafanaID, err := d.startGrafana(ctx)
	if err != nil {
		return err
	}
	defer d.removeContainer(grafanaID, grafanaName)

	// Take screenshot
	if err := d.takeScreenshot(ctx); err != nil {
		return err
	}
	return nil
}

func (d *Dashboard) removeNetwork() {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout*time.Second)
	defer cancel()

	d.log.WithField("name", cmdNetworkName).Info("Removing network")

	if err := d.cli.NetworkRemove(ctx, cmdNetworkName); err != nil {
		d.log.Error(err.Error())
	}
}

func (d *Dashboard) createNetwork(ctx context.Context) error {
	d.log.WithField("name", cmdNetworkName).Info("Creating network")
	resp, err := d.cli.NetworkCreate(ctx, cmdNetworkName, types.NetworkCreate{
		Driver:         "bridge",
		CheckDuplicate: true,
	})

	if err != nil {
		return err
	}

	if resp.Warning != "" {
		d.log.Warn(resp.Warning)
	}

	return nil
}

func (d *Dashboard) removeContainer(id, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout*time.Second)
	defer cancel()
	d.log.WithFields(logrus.Fields{
		"id":   id,
		"name": name,
	}).Info("Stopping container")

	if err := d.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true}); err != nil {
		d.log.Warn(err.Error())
	}
}

func (d *Dashboard) startGrafana(ctx context.Context) (string, error) {
	d.log.WithFields(logrus.Fields{
		"name":  grafanaName,
		"image": grafanaDockerImg,
	}).Info("Creating container")
	resp, err := d.cli.ContainerCreate(ctx, &container.Config{
		Image: grafanaDockerImg,
	}, nil, nil, nil, grafanaName)

	if err != nil {
		return "", err
	}

	if err := d.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	d.log.WithFields(logrus.Fields{
		"id":  resp.ID,
		"img": grafanaDockerImg,
	}).Info("Connecting to network")

	if err := d.cli.NetworkConnect(ctx, cmdNetworkName, resp.ID, nil); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (d *Dashboard) takeScreenshot(ctx context.Context) error {
	hostConfig := container.HostConfig{
		CapAdd: []string{"SYS_ADMIN"},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/tmp",
				Target: "/working",
			},
		},
	}

	d.log.WithFields(logrus.Fields{
		"image": pupeteerDockerImg,
		"name":  pupeteerName,
		"cmd":   puppeteerCmd,
	}).Info("Creating container")
	resp, err := d.cli.ContainerCreate(ctx, &container.Config{
		Image:      pupeteerDockerImg,
		Cmd:        puppeteerCmd,
		WorkingDir: "/working",
	}, &hostConfig, nil, nil, pupeteerName)
	if err != nil {
		return err
	}

	defer d.removeContainer(resp.ID, pupeteerName)

	d.log.WithFields(logrus.Fields{
		"id":   resp.ID,
		"img":  pupeteerDockerImg,
		"name": cmdNetworkName,
	}).Info("Connecting to network")
	if err := d.cli.NetworkConnect(ctx, cmdNetworkName, resp.ID, nil); err != nil {
		return err
	}

	d.log.WithFields(logrus.Fields{
		"id":   resp.ID,
		"name": pupeteerName,
	}).Info("Starting container")
	if err := d.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := d.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			d.log.WithFields(logrus.Fields{
				"id":         resp.ID,
				"img":        pupeteerDockerImg,
				"name":       cmdNetworkName,
				"statusCode": status.StatusCode,
			}).Error("Container completed with a non-zero status code!")
			d.log.Info(d.fetchContainerLogs(resp.ID))
			return errors.New("non-zero return code")
		}

	case <-ctx.Done():
		d.log.WithFields(logrus.Fields{
			"name": pupeteerName,
		}).Error("Timedout waiting for container")

		d.log.Info(d.fetchContainerLogs(resp.ID))

		return errors.New("timedout waiting for screenshot")
	}

	return nil
}

func (d *Dashboard) fetchContainerLogs(id string) io.Writer {
	d.log.Info("Fetching logs ...")
	ctxLogs, cancelLogs := context.WithTimeout(context.Background(), cleanupTimeout*time.Second)
	defer cancelLogs()
	ioCloser, err := d.cli.ContainerLogs(
		ctxLogs,
		id,
		types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true},
	)
	if err != nil {
		panic(err)
	}

	buf := new(strings.Builder)
	_, _ = stdcopy.StdCopy(buf, buf, ioCloser)
	return buf
}
