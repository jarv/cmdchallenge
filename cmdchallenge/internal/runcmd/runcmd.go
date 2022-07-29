package runcmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Result struct {
	Output     string
	ExitCode   int
	Correct    *bool  `json:",omitempty"`
	OutputPass *bool  `json:",omitempty"`
	TestPass   string `json:",omitempty"`
	Error      string
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func runCombinedOutput(command string) (cmdout string, exitCode int) {
	bashArgs := []string{"-O", "globstar", "-c", "export MANPAGER=cat;" + command}
	shArgs := []string{"-c", command}
	var args *[]string
	var interpreter string

	if fileExists("/bin/bash") {
		interpreter = "bash"
		args = &bashArgs
	} else {
		interpreter = "sh"
		args = &shArgs
	}

	cmd := exec.Command(interpreter, *args...)
	outb, err := cmd.CombinedOutput()
	cmdout = fmt.Sprintf("%s", outb)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		}
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}
	return cmdout, exitCode
}
