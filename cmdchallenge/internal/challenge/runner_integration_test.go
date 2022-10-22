package challenge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/require"
	"gitlab.com/jarv/cmdchallenge/internal/config"
)

const (
	jsonExt string = ".json"
)

func chSlugs(t *testing.T, path string) []os.DirEntry {
	items, err := os.ReadDir(path)
	require.NoError(t, err)
	return items
}

func TestChallengesExpectPass(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ass := require.New(t)
	req := require.New(t)
	cfg := config.New()

	for _, item := range chSlugs(t, cfg.ChallengePath()) {
		item := item
		if filepath.Ext(item.Name()) != jsonExt {
			continue
		}

		slug := strings.TrimSuffix(filepath.Base(item.Name()), filepath.Ext(item.Name()))

		chJSON, err := cfg.JSONForSlug(slug)
		req.NoError(err)

		ch, err := NewChallenge(chJSON)
		req.NoError(err)

		t.Run(ch.Slug(), func(t *testing.T) {
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
	cfg := config.New()

	for _, item := range chSlugs(t, cfg.ChallengePath()) {
		item := item
		if filepath.Ext(item.Name()) != jsonExt {
			continue
		}

		slug := strings.TrimSuffix(filepath.Base(item.Name()), filepath.Ext(item.Name()))

		chJSON, err := cfg.JSONForSlug(slug)
		req.NoError(err)

		ch, err := NewChallenge(chJSON)
		req.NoError(err)

		t.Run(ch.Slug(), func(t *testing.T) {
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
