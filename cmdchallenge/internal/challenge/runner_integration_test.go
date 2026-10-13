package challenge

import (
	"log/slog"
	"testing"

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
			runner := NewRunner(slog.Default(), cfg)
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
			runner := NewRunner(slog.Default(), cfg)
			for _, failure := range ch.ExpectedFailures() {
				result, err := runner.RunContainer(failure, ch)
				req.NoError(err)
				ass.NotNil(result.Correct)
				ass.False(*result.Correct)
			}
		})
	}
}

func TestRunnerManPages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker integration test in short mode.")
	}
	runner := NewRunner(slog.Default(), config.New(config.ConfigOpts{}))
	ch := helloWorldCh(t)
	for _, command := range []string{"man grep", "man ls", "man bash", "man awk"} {
		t.Run(command, func(t *testing.T) {
			result, err := runner.RunContainer(command, ch)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Nil(t, result.ErrorInternal)
			require.NotNil(t, result.ExitCode)
			require.Equal(t, 0, *result.ExitCode)
			require.NotNil(t, result.Output)
			require.Contains(t, *result.Output, "NAME")
			require.Contains(t, *result.Output, "SYNOPSIS")
			require.NotContains(t, *result.Output, "This system has been minimized")
		})
	}
}
