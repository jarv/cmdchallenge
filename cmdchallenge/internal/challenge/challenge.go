package challenge

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/google/shlex"
)

const (
	DefaultImg string = "cmd"
)

type ChInfo struct {
	Slug           *string `json:"slug,omitempty"`
	Version        *int    `json:"version,omitempty"`
	Dir            *string `json:"dir,omitempty"`
	Img            *string `json:"img,omitempty"`
	Example        *string `json:"example,omitempty"`
	ExpectedOutput *struct {
		Lines *[]string `json:"lines,omitempty"`
	} `json:"expected_output,omitempty"`
	ExpectedFailures *[]string `json:"expected_failures,omitempty"`
}

type Challenge struct {
	chJSON []byte
	chInfo *ChInfo
}

func New(chFile string) (*Challenge, error) {
	chJSON, err := ioutil.ReadFile(chFile)
	if err != nil {
		return nil, err
	}

	var chInfo ChInfo

	if err := json.Unmarshal(chJSON, &chInfo); err != nil {
		return nil, err
	}

	return &Challenge{
		chJSON: chJSON,
		chInfo: &chInfo,
	}, nil
}

func (c *Challenge) HasExpectedOutput() bool {
	if c.chInfo.ExpectedOutput == nil || c.chInfo.ExpectedOutput.Lines == nil {
		return false
	}
	return true
}

func (c *Challenge) ExpectedOutput() []string {
	return *c.chInfo.ExpectedOutput.Lines
}

func (c *Challenge) ExpectedFailures() []string {
	if c.chInfo.ExpectedFailures == nil {
		return []string{}
	}
	return *c.chInfo.ExpectedFailures
}

func (c *Challenge) Example() string {
	return *c.chInfo.Example
}

func (c *Challenge) Slug() string {
	return *c.chInfo.Slug
}

func (c *Challenge) Version() int {
	return *c.chInfo.Version
}

func (c *Challenge) Dir() string {
	if c.chInfo.Dir == nil {
		return *c.chInfo.Slug
	}

	return *c.chInfo.Dir
}

func (c *Challenge) Img() string {
	if c.chInfo.Img == nil {
		return DefaultImg
	}

	return *c.chInfo.Img
}

func (c *Challenge) Fingerprint(cmd string) (string, error) {
	cmdShlex, err := shlex.Split(cmd)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256([]byte(string(c.chJSON) + strings.Join(cmdShlex, " ")))
	return fmt.Sprintf("%x", sum), nil
}
