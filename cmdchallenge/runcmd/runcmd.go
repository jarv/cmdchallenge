package main

/*
 * parse challenges.json, print all keys
 * get output for keys
 * split string
 *
 */
import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const defaultFailedCode = 1

func check(e error) {
	if e != nil {
		errorExit(e.Error())
	}
}

func errorExit(s string) {
	type CmdError struct{ Error string }
	result := CmdError{Error: s}
	b, _ := json.MarshalIndent(result, "", "  ")
	os.Stdout.Write(b)
	os.Exit(0)
}

func runCombinedOutput(command string) (cmdout string, exitCode int) {
	cmd := exec.Command("bash", "-O", "globstar", "-c", "export MANPAGER=cat;"+command)
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

func newBool(b bool) *bool {
	return &b
}

func main() {
	progDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	check(err)

	type CmdResult struct {
		CmdOut                  string
		CmdExitCode             int
		OutputPass              *bool  `json:",omitempty"`
		TestsPass               *bool  `json:",omitempty"`
		TestsOut                string `json:",omitempty"`
		AfterRandOutputPass     *bool  `json:",omitempty"`
		AfterRandExpectedOutput string `json:",omitempty"`
		AfterRandOutput         string `json:",omitempty"`
		AfterRandTestsPass      *bool  `json:",omitempty"`
		AfterRandTestsOut       string `json:",omitempty"`
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s <command>:\n", os.Args[0])
		flag.PrintDefaults()
	}

	slugNamePtr := flag.String("slug", "hello_world", "a string")
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}
	commandArg := flag.Args()[0]
	// if the base64 decode fails assume the command was passed in
	// without encoding
	decoded, err := base64.StdEncoding.DecodeString(commandArg)
	var command string
	if err != nil {
		command = commandArg
	} else {
		command = string(decoded)
	}

	challengesJSON, err := ioutil.ReadFile(progDir + "/ch/" + *slugNamePtr + ".json")
	if err != nil {
		errorExit("Challenge " + *slugNamePtr + " not found.")
	}
	ch, err := readChallenge(challengesJSON)
	check(err)
	cmdOut, cmdExitCode := runCombinedOutput(command)

	result := CmdResult{
		CmdOut:      cmdOut,
		CmdExitCode: cmdExitCode,
	}

	if ch.HasExpectedOutput() {
		result.OutputPass = newBool(ch.MatchesOutput(cmdOut))
	}

	testFile := progDir + "/cmdtests/" + ch.Slug
	if _, err := os.Stat(testFile); err == nil {
		testOut, testExitCode := runCombinedOutput(testFile)
		if testExitCode == 0 {
			result.TestsPass = newBool(true)
		} else {
			result.TestsPass = newBool(false)
		}
		result.TestsOut = testOut
	}

	randFile := progDir + "/randomizers/" + ch.Slug
	if _, err := os.Stat(randFile); err == nil {
		randOut, randExitCode := runCombinedOutput(randFile)
		result.AfterRandExpectedOutput = randOut
		if randExitCode == 0 {
			// Randomizer has run, check output against the new
			// expected output
			newCh := challenge{}
			lineSlice := strings.Split(randOut, "\n")
			lineSlice = lineSlice[:len(lineSlice)-1]
			newCh.ExpectedOutput.Lines = lineSlice
			newCh.ExpectedOutput.Order = ch.ExpectedOutput.Order
			newCh.ExpectedOutput.ReSub = ch.ExpectedOutput.ReSub
			// TODO exit code?
			randCmdOut, _ := runCombinedOutput(command)
			result.AfterRandOutput = randCmdOut
			result.AfterRandOutputPass = newBool(newCh.MatchesOutput(randCmdOut))
			if _, err := os.Stat(testFile); err == nil {
				randTestOut, randTestExitCode := runCombinedOutput(testFile)
				if randTestExitCode == 0 {
					result.AfterRandTestsPass = newBool(true)
				} else {
					result.AfterRandTestsPass = newBool(false)
				}
				result.AfterRandTestsOut = randTestOut
			}
		} else {
			errorExit("Unable to run random check.")
		}
	}
	b, err := json.Marshal(result)
	check(err)
	os.Stdout.Write(append(b, "\n"...))
}
