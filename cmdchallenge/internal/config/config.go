package config

import (
	_ "embed"
	"errors"
	"runtime"
	"time"
)

const oopsBin = "oops-this-will-delete-bin-dirs"
const devTagSuffix = "-testing"

var (
	ErrInvalidRegistryImgURI = errors.New("registry image doesn't exist")
)

type ConfigOpts struct {
	DevMode       bool
	RateLimit     bool
	DevTag        bool
	DBFile        string
	StaticDistDir string
}

type Config struct {
	CmdTimeout           time.Duration
	RateLimit            bool
	RunCmdRegistryImgTag string
	RunCmdRegistryImg    string
	RegistryAuth         string
	RunCmdTimeout        time.Duration
	RemoveImageTimeout   time.Duration
	PullImageTimeout     time.Duration
	DBFile               string
	DevMode              bool
	CMDImgNames          []string
	OopsBin              string
	SolutionsKeyPrefix   string
	Caller               string
	registryImgURIs      map[string]string
	StaticDistDir        string
}

func New(c ConfigOpts) *Config {
	tagSuffix := ""
	if c.DevTag {
		tagSuffix = devTagSuffix
	}

	return &Config{
		CmdTimeout:         5 * time.Second,
		RegistryAuth:       "",
		RunCmdTimeout:      6 * time.Second,
		PullImageTimeout:   5 * time.Minute,
		RateLimit:          c.RateLimit,
		RemoveImageTimeout: 60 * time.Second,
		DevMode:            c.DevMode,
		DBFile:             c.DBFile,
		OopsBin:            oopsBin,
		SolutionsKeyPrefix: "s/solutions",
		StaticDistDir:      c.StaticDistDir,

		registryImgURIs: map[string]string{
			"cmd":        "cmd:" + runtime.GOARCH + tagSuffix,
			"cmd-no-bin": "cmd-no-bin:" + runtime.GOARCH + tagSuffix,
		},
	}
}

func (c *Config) RegistryImgURI(name string) (string, error) {
	if val, ok := c.registryImgURIs[name]; ok {
		return val, nil
	}
	return "", ErrInvalidRegistryImgURI
}
