package runcmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"gitlab.com/jarv/cmdchallenge/internal/challenge"
	"gitlab.com/jarv/cmdchallenge/internal/config"
)

const (
	oopsPid = 42
)

var ErrUnableToSetPidOops = errors.New("unable to set pid for oops process")
var ErrTimeout = errors.New("timed out executing command")
var ErrCombinedOutput = errors.New("error getting combined output")

type RunCmd struct {
	log     *slog.Logger
	config  *config.Config
	oopsCmd *exec.Cmd
}

func New(log *slog.Logger, cfg *config.Config) *RunCmd {
	return &RunCmd{log, cfg, nil}
}

func (r *RunCmd) startOops(ctx context.Context) (*exec.Cmd, chan string, error) {
	for i := 0; i < oopsPid; i++ {
		cmd := exec.Command("bash", "--help") // #nosec G204
		if err := cmd.Start(); err != nil {
			return nil, nil, err
		}
		if err := cmd.Wait(); err != nil {
			return nil, nil, err
		}
		r.log.Info(fmt.Sprintf("process pid: cmd.Process.Pid=%v oopsPid=%v", cmd.Process.Pid, oopsPid))
		if cmd.Process.Pid >= (oopsPid - 1) {
			break
		}
	}

	cmd := exec.CommandContext(ctx, r.config.OopsBin) // #nosec G204
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	if cmd.Process.Pid != oopsPid {
		r.log.Info(fmt.Sprintf("Unable to set process pid: cmd.Process.Pid=%v oopsPid=%v", cmd.Process.Pid, oopsPid))
		return nil, nil, ErrUnableToSetPidOops
	}
	oopsDone := make(chan string)
	go func() {
		_ = cmd.Wait()
		oopsDone <- "done!"
	}()
	return cmd, oopsDone, nil
}

func (r *RunCmd) stopOops(oops *exec.Cmd) {
	if err := oops.Process.Kill(); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return
		}
		r.log.Error("unable to kill oops process", "err", err)
	}
}

func (r *RunCmd) Run(ch *challenge.Challenge, command string) string {
	slug := ch.Slug()
	resp := &challenge.CmdResponse{Correct: toPtr(true)}
	var oopsDone chan string
	ctx, cancel := context.WithTimeout(context.Background(), r.config.CmdTimeout)
	defer cancel()

	// For some challenges, start the oops process
	if strings.HasPrefix(slug, "oops") {
		var err error
		var oopsCmd *exec.Cmd
		oopsCmd, oopsDone, err = r.startOops(ctx)
		if err != nil {
			return r.marshalIncorrectErrInt(err, resp, "error starting process for oops challenge")
		}
		defer r.stopOops(oopsCmd)
	}

	// Run command and record output and exit code
	cmdOut, exitCode, err := r.runCombinedOutput(ctx, command)
	if err != nil {
		return r.marshalIncorrectErrInt(err, resp, err.Error())
	}
	resp.Output = cmdOut
	resp.ExitCode = exitCode

	// Check against expected lines if specified
	if ch.HasExpectedLines() {
		okM, err := ch.MatchesLines(*resp.Output, nil)
		if err != nil {
			return r.marshalIncorrectErrInt(err, resp, "Unexpected error when checking for expected lines")
		}
		if !okM {
			return r.marshalIncorrectErr(resp, "Output does not match expected lines")
		}

		if ch.HasRandomizer() {
			okR, err := r.checkAfterRandomizer(ctx, ch, command)
			if err != nil {
				return r.marshalIncorrectErrInt(err, resp, err.Error())
			}
			if !okR {
				return r.marshalIncorrectErr(resp, "Output does not match expected lines after randomizing data")
			}
		}
	}

	// Run extra checks if they are specified
	if ch.HasCheck() {
		checkResult, err := challenge.NewCheck(r.log, ch, oopsDone).RunCheck()
		if err != nil {
			return r.marshalIncorrectErrInt(err, resp, "Unexpected error when running checks")
		}

		if checkResult != "" {
			return r.marshalIncorrectErr(resp, checkResult)
		}
	}

	return marshalOrPanic(resp)
}

func (r *RunCmd) checkAfterRandomizer(ctx context.Context, ch *challenge.Challenge, command string) (bool, error) {
	rnd := challenge.NewRandomizer(r.log, ch)
	rndExpectedLines, err := rnd.RunRandomizer()
	if err != nil {
		return false, errors.New("unexpected error when randomizing data")
	}

	// Run command after randomizer
	outAfterRnd, _, err := r.runCombinedOutput(ctx, command)
	if err != nil {
		return false, err
	}

	c, err := ch.MatchesLines(*outAfterRnd, &rndExpectedLines)
	if err != nil {
		return false, err
	}

	return c, nil
}

func updateCorrect(resp *challenge.CmdResponse, correct bool) {
	if resp.Correct != nil {
		resp.Correct = toPtr(*resp.Correct && correct)
	} else {
		resp.Correct = toPtr(correct)
	}
}

func (r *RunCmd) marshalIncorrectErrInt(err error, resp *challenge.CmdResponse, errMsg string) string {
	resp.ErrorInternal = toPtr(errMsg)
	// Internal errors are logged
	r.log.Error(*resp.ErrorInternal, "err", err)
	updateCorrect(resp, false)
	return marshalOrPanic(resp)
}

func (r *RunCmd) marshalIncorrectErr(resp *challenge.CmdResponse, errMsg string) string {
	resp.Error = toPtr(errMsg)
	if resp.ExitCode == nil {
		resp.ExitCode = toPtr(-1)
	}

	updateCorrect(resp, false)
	return marshalOrPanic(resp)
}

func marshalOrPanic(resp *challenge.CmdResponse) string {
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	return string(jsonResp)
}

//nolint:gocritic // unnamedResult
func (r *RunCmd) runCombinedOutput(ctx context.Context, command string) (*string, *int, error) {
	bashArgs := []string{"-O", "globstar", "-c", "export MANPAGER=cat;" + command}
	cmd := exec.Command("bash", bashArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	type cmdResult struct {
		outb []byte
		err  error
	}
	cmdDone := make(chan cmdResult, 1)
	go func() {
		outb, cmderr := cmd.CombinedOutput()
		cmdDone <- cmdResult{outb, cmderr}
	}()

	select {
	// Wait for the process to finish or kill it after a timeout (whichever happens first)
	case <-ctx.Done():
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		return nil, nil, ErrTimeout
	case result := <-cmdDone:
		var exerr *exec.ExitError
		exitCode := 0

		if errors.As(result.err, &exerr) {
			exitCode = exerr.ExitCode()
		} else if result.err != nil {
			return nil, nil, result.err
		}
		return toPtr(string(result.outb)), toPtr(exitCode), nil
	}
}

type ptrConvert interface {
	string | bool | int
}

func toPtr[T ptrConvert](i T) *T {
	return &i
}
