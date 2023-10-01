package challenge

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

const (
	DefaultImg    string = "cmd"
	reSubElements int    = 2 // number of elements expected for reSub in yml config
)

type ChInfo struct {
	Slug           *string `yaml:"slug,omitempty"`
	Version        *int    `yaml:"version,omitempty"`
	Dir            *string `yaml:"dir,omitempty"`
	Img            *string `yaml:"img,omitempty"`
	Example        *string `yaml:"example,omitempty"`
	ExpectedOutput *struct {
		Order             *bool     `yaml:"order,omitempty"`
		IgnoreNonMatching *bool     `yaml:"ignore_non_matching,omitempty"`
		ReSub             *[]string `yaml:"re_sub,omitempty"`
		Lines             *[]string `yaml:"lines,omitempty"`
	} `yaml:"expected_output,omitempty"`
	ExpectedFailures *[]string `yaml:"expected_failures,omitempty"`
}

type Challenge struct {
	chInfo *ChInfo
}

//go:embed challenges.yaml
var challengesYAML string

type ChallengeOptions struct {
	Slug           string
	ChallengesYAML string
}

func NewChallenge(opt ChallengeOptions) (*Challenge, error) {
	var challenges []ChInfo

	if opt.ChallengesYAML == "" {
		opt.ChallengesYAML = challengesYAML
	}

	if err := yaml.Unmarshal([]byte(opt.ChallengesYAML), &challenges); err != nil {
		return nil, fmt.Errorf("%s\n%w", opt.ChallengesYAML, err)
	}

	for _, c := range challenges {
		c := c
		if *c.Slug == opt.Slug {
			return &Challenge{
				chInfo: &c,
			}, nil
		}
	}
	return nil, fmt.Errorf("unable to find challenge for slug %s", opt.Slug)
}

func (c *Challenge) HasExpectedLines() bool {
	if c.chInfo.ExpectedOutput == nil || c.chInfo.ExpectedOutput.Lines == nil {
		return false
	}
	return true
}

func (c *Challenge) ExpectedLines() []string {
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

func (c *Challenge) HasOrderedExpectedLines() bool {
	if c.chInfo.ExpectedOutput.Order == nil {
		return true
	} else {
		return *c.chInfo.ExpectedOutput.Order
	}
}

func (c *Challenge) HasIgnoreNonMatching() bool {
	if c.chInfo.ExpectedOutput.IgnoreNonMatching == nil {
		return false
	} else {
		return *c.chInfo.ExpectedOutput.IgnoreNonMatching
	}
}

func (c *Challenge) Img() string {
	if c.chInfo.Img == nil {
		return DefaultImg
	}

	return *c.chInfo.Img
}

func (c *Challenge) MatchesLines(cmdOut string, l *[]string) (bool, error) {
	// Remove leading and trailing spaces from cmdOut
	lines := strings.Split(strings.TrimSpace(cmdOut), "\n")
	var expectedLines *[]string

	if l != nil {
		expectedLines = l
	} else {
		expectedLines = c.chInfo.ExpectedOutput.Lines
	}

	if c.chInfo.ExpectedOutput.ReSub != nil {
		if len(*c.chInfo.ExpectedOutput.ReSub) != reSubElements {
			return false, errors.New("re_sub should have two elements")
		}
		r, err := regexp.Compile((*c.chInfo.ExpectedOutput.ReSub)[0])
		if err != nil {
			return false, errors.New("unable to compile re_sub regex")
		}

		for i := range lines {
			lines[i] = r.ReplaceAllString(lines[i], (*c.chInfo.ExpectedOutput.ReSub)[1])
		}
	}

	if !c.HasOrderedExpectedLines() {
		// Order doesn't matter, sort before comparing
		sort.Strings(*expectedLines)
		sort.Strings(lines)
	}

	if c.HasIgnoreNonMatching() {
		lines = removeNonMatching(lines, *expectedLines)
	}

	return cmp.Equal(lines, *expectedLines), nil
}

func (c *Challenge) HasCheck() bool {
	_, exists := checkTable[c.Slug()]

	return exists
}

func (c *Challenge) HasRandomizer() bool {
	_, exists := rndTable[c.Slug()]

	return c.HasExpectedLines() && exists
}

func removeNonMatching(lines, expectedLines []string) []string {
	matchingLines := []string{}

	for _, l := range lines {
		match := false
		for _, e := range expectedLines {
			if e == l {
				match = true
				break
			}
		}

		if match {
			matchingLines = append(matchingLines, l)
		}
	}
	return matchingLines
}
