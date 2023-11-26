package challenge

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"gitlab.com/jarv/cmdchallenge/internal/config"
)

const (
	BaseWorkingDir string = "/var/challenges"
)

type RunnerExecutor interface {
	PullImages() error
	RunContainer(cmd string, ch *Challenge) (*CmdResponse, error)
}

type Runner struct {
	log *slog.Logger
	cfg *config.Config
	cli *client.Client
}

func NewRunner(log *slog.Logger, cfg *config.Config) *Runner {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	r := Runner{log, cfg, cli}
	return &r
}

func (r *Runner) PullImages() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.cfg.PullImageTimeout)
	defer cancel()
	opts := &types.ImagePullOptions{
		// RegistryAuth: r.cfg.RegistryAuth,
	}

	for _, imgName := range r.cfg.CMDImgNames {
		imgURI, err := r.cfg.RegistryImgURI(imgName)
		if err != nil {
			return err
		}

		r.log.Info("Starting image pull", "imgURI", imgURI)
		_, err = r.cli.ImagePull(ctx, imgURI, *opts)
		if err != nil {
			return err
		}
		r.log.Info("Finished image pull", "imgURI", imgURI)
	}

	return nil
}

func (r *Runner) containerLogs(ctx context.Context, id string, showStdout, showStderr bool) string {
	ioCloser, err := r.cli.ContainerLogs(
		ctx,
		id,
		types.ContainerLogsOptions{ShowStdout: showStdout, ShowStderr: showStderr},
	)
	if err != nil {
		panic(err)
	}
	defer ioCloser.Close()

	buf := new(strings.Builder)
	_, _ = stdcopy.StdCopy(buf, buf, ioCloser)

	return buf.String()
}

func (r *Runner) RunContainer(cmd string, ch *Challenge) (*CmdResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.cfg.RunCmdTimeout)
	defer cancel()

	workingDir := path.Join(BaseWorkingDir, ch.Dir())

	hostConfig := container.HostConfig{
		NetworkMode: "none",
		Resources:   container.Resources{Memory: 10e+7},
	}

	runCmd := []string{
		"runcmd",
		"-cmd",
		"-slug",
		ch.Slug(),
		base64.StdEncoding.EncodeToString([]byte(cmd)),
	}

	registryImgURI, err := r.cfg.RegistryImgURI(ch.Img())
	if err != nil {
		return nil, err
	}

	r.log.Info("Creating container", "Image", registryImgURI, "Cmd", runCmd, "WorkingDir", workingDir)
	resp, err := r.cli.ContainerCreate(ctx, &container.Config{
		Image:      registryImgURI,
		Cmd:        runCmd,
		WorkingDir: workingDir,
	}, &hostConfig, nil, nil, "")
	if err != nil {
		// Error response from daemon: No such image: registry.gitlab.com/jarv/cmdchallenge/cmd:latest
		return nil, err
	}

	if err := r.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	statusCh, errCh := r.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case status := <-statusCh:
		stdout := r.containerLogs(ctx, resp.ID, true, false)
		stderr := r.containerLogs(ctx, resp.ID, false, true)

		if stderr != "" {
			r.log.Error("Container logs:\n" + "--------------\n" + stderr + "--------------")
		}

		r.log.Info("Got response from runner",
			"cmd", cmd, "slug", ch.Slug(), "workingDir", workingDir,
			"statusCode", status.StatusCode,
			"stdout", stdout)

		if status.StatusCode != 0 {
			r.log.Error("Container completed with a non-zero status code!",
				"statusCode", status.StatusCode,
			)
			return nil, ErrRunnerNonZeroReturn
		}

		var cmdResponse CmdResponse

		if err := json.Unmarshal([]byte(stdout), &cmdResponse); err != nil {
			r.log.Error("Unable to decode result", "stdout", stdout)
			return nil, ErrRunnerDecodeResult
		}

		return &cmdResponse, nil

	case <-ctx.Done():
		r.log.Info("Removing container due to timeout", "Image", registryImgURI, "id", resp.ID, "slug", ch.Slug())
		if err := r.removeImage(resp.ID); err != nil {
			return nil, err
		}
		return nil, ErrRunnerTimeout
	}
	return nil, nil
}

func (r *Runner) removeImage(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.cfg.RemoveImageTimeout)
	defer cancel()
	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	if err := r.cli.ContainerRemove(ctx, id, removeOptions); err != nil {
		r.log.Error("Unable to remove container", "err", err)
		return err
	}
	return nil
}
