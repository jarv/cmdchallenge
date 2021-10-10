package runner

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	// "github.com/gdexlab/go-render/render"
	"github.com/sirupsen/logrus"
	"gitlab.com/jarv/cmdchallenge/internal/challenge"
	"gitlab.com/jarv/cmdchallenge/internal/config"
)

var (
	ErrNonZeroReturn     = errors.New("non-zero return code")
	ErrDecodeResult      = errors.New("unable to decode result")
	ErrResultNotFound    = errors.New("result not found")
	ErrTimeout           = errors.New("runner timeout")
	ErrImgRemovalTimeout = errors.New("unable to cleanup after timeout")
)

const (
	BaseWorkingDir string = "/var/challenges"
)

type RunnerExecutor interface {
	PullImages() error
	RunContainer(ch *challenge.Challenge, cmd string) (*RunnerResult, error)
}

type RunnerResultStorer interface {
	GetResult(fingerprint string) (*RunnerResult, error)
	CreateResult(fingerprint, cmd, slug string, version int, result *RunnerResult) error
	IncrementResult(fingerprint string) error
	TopCmdsForSlug(slug string) ([]string, error)
}

type Runner struct {
	log *logrus.Logger
	cfg *config.Config
	cli *client.Client
}

type RunnerResult struct {
	Output              *string `json:",omitempty"`
	Cmd                 *string `json:",omitempty"`
	ExitCode            *int32  `json:",omitempty"`
	Correct             *bool   `json:",omitempty"`
	OutputPass          *bool   `json:",omitempty"`
	TestPass            *bool   `json:",omitempty"`
	AfterRandOutputPass *bool   `json:",omitempty"`
	AfterRandTestPass   *bool   `json:",omitempty"`
	Error               *string `json:",omitempty"`
}

func New(log *logrus.Logger, cfg *config.Config) *Runner {
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

	for _, imgName := range r.cfg.CMDImageNames {
		imgURI := r.cfg.RegistryImgURI(imgName)
		r.log.Infof("Pulling %s", imgURI)
		reader, err := r.cli.ImagePull(ctx, imgURI, *opts)
		if err != nil {
			return err
		}
		_, _ = io.Copy(r.log.Writer(), reader)
	}

	return nil
}

func (r *Runner) RunContainer(ch *challenge.Challenge, cmd string) (*RunnerResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.cfg.RunCmdTimeout)
	defer cancel()

	workingDir := path.Join(BaseWorkingDir, ch.Dir())

	hostConfig := container.HostConfig{
		NetworkMode: "none",
		Resources:   container.Resources{Memory: 10e+7},
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   r.cfg.ROVolumeDir,
				Target:   "/ro_volume",
				ReadOnly: true,
			},
		},
	}

	runCmd := []string{
		"runcmd",
		"--slug",
		ch.Slug(),
		base64.StdEncoding.EncodeToString([]byte(cmd)),
	}

	resp, err := r.cli.ContainerCreate(ctx, &container.Config{
		Image:      r.cfg.RegistryImgURI(ch.Img()),
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
		ioCloser, err := r.cli.ContainerLogs(
			ctx,
			resp.ID,
			types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true},
		)
		if err != nil {
			panic(err)
		}

		buf := new(strings.Builder)
		_, _ = stdcopy.StdCopy(buf, buf, ioCloser)

		output := buf.String()
		r.log.WithFields(logrus.Fields{"cmd": cmd, "slug": ch.Slug(), "workingDir": workingDir,
			"statusCode": status.StatusCode,
			"output":     output,
		}).Info("Got response from runner")

		if status.StatusCode != 0 {
			r.log.WithFields(logrus.Fields{"cmd": cmd, "slug": ch.Slug(), "workingDir": workingDir,
				"statusCode": status.StatusCode,
				"output":     output,
			}).Error("Container completed with a non-zero status code!")
			return nil, ErrNonZeroReturn
		}

		var runnerResult RunnerResult

		if err := json.Unmarshal([]byte(output), &runnerResult); err != nil {
			r.log.WithFields(logrus.Fields{"cmd": cmd, "slug": ch.Slug(), "workingDir": workingDir,
				"statusCode": status.StatusCode,
				"output":     output,
			}).Error("Unable to decode result")
			return nil, ErrDecodeResult
		}

		return &runnerResult, nil

	case <-ctx.Done():
		r.log.WithFields(logrus.Fields{"cmd": cmd, "slug": ch.Slug(), "workingDir": workingDir,
			"id": resp.ID,
		}).Warn("Removing container due to timeout")

		if err := r.removeImage(resp.ID); err != nil {
			return nil, err
		}
		return nil, ErrTimeout
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
		r.log.Errorf("Unable to remove container: %s", err)
		return err
	}
	return nil
}
