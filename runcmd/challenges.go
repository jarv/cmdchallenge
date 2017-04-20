package main

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

func readChallenge(cBlob []byte) (challenge, error) {
	ch := challenge{
		ExpectedOutput: expectedOutput{
			Order: true,
		},
	}
	er := json.Unmarshal(cBlob, &ch)
	return ch, er
}

type expectedOutput struct {
	Lines []string `json:"lines"`
	Order bool     `json:"order"`
	ReSub []string `json:"re_sub"`
}

type challenge struct {
	Slug           string         `json:"slug"`
	Version        int            `json:"version"`
	Author         string         `json:"author"`
	Example        string         `json:"example"`
	ExpectedOutput expectedOutput `json:"expected_output"`
}

func (ch challenge) HasExpectedOutput() bool {
	return len(ch.ExpectedOutput.Lines) > 0
}

func (ch challenge) MatchesOutput(cmdOut string) bool {
	if !ch.HasExpectedOutput() {
		return true
	}
	lineSlice := strings.Split(cmdOut, "\n")

	// Always assume that cmdOut ends with a "\n"
	lineSlice = lineSlice[:len(lineSlice)-1]

	if len(ch.ExpectedOutput.ReSub) != 0 {
		r, err := regexp.Compile(ch.ExpectedOutput.ReSub[0])
		check(err)
		for i := range lineSlice {
			lineSlice[i] = r.ReplaceAllString(lineSlice[i], ch.ExpectedOutput.ReSub[1])
		}
	}

	if ch.ExpectedOutput.Order {
		return testEqual(ch.ExpectedOutput.Lines[:], lineSlice)
	}
	// Order doesn't matter, sort before comparing
	sortSlice(ch.ExpectedOutput.Lines[:])
	sortSlice(lineSlice)
	return testEqual(ch.ExpectedOutput.Lines[:], lineSlice)
}

func sortSlice(a []string) {
	sort.Slice(a,
		func(i, j int) bool { return a[i] < a[j] })
}

func testEqual(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
