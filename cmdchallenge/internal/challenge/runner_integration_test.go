package challenge

import (
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/require"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gopkg.in/yaml.v3"
)

func chSlugs(t *testing.T) []string {
	var challenges []ChInfo
	var slugs []string

	err := yaml.Unmarshal([]byte(challengesYAML), &challenges)
	require.NoError(t, err)

	for _, c := range challenges {
		slugs = append(slugs, *c.Slug)
	}
	return slugs
}

func TestChallengesExpectPass(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ass := require.New(t)
	req := require.New(t)
	cfg := config.New(config.ConfigOpts{})

	for _, slug := range chSlugs(t) {
		slug := slug

		ch, err := NewChallenge(ChallengeOptions{Slug: slug})
		req.NoError(err)

		t.Run(slug, func(t *testing.T) {
			t.Parallel()
			runner := NewRunner(testr.New(t), cfg)
			result, err := runner.RunContainer(ch.Example(), ch)
			ass.NoError(err)
			ass.NotNil(result.Correct)
			ass.True(*result.Correct)
			ass.Nil(result.Error)
		})
	}
}

func TestChallengesExpectFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ass := require.New(t)
	req := require.New(t)
	cfg := config.New(config.ConfigOpts{})

	for _, slug := range chSlugs(t) {
		slug := slug

		ch, err := NewChallenge(ChallengeOptions{Slug: slug})
		req.NoError(err)

		t.Run(slug, func(t *testing.T) {
			t.Parallel()
			runner := NewRunner(testr.New(t), cfg)
			for _, failure := range ch.ExpectedFailures() {
				result, err := runner.RunContainer(failure, ch)
				req.NoError(err)
				ass.NotNil(result.Correct)
				ass.False(*result.Correct)
			}
		})
	}
}
