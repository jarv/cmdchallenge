package runner

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"testing"
	"time"

	"gitlab.com/jarv/cmdchallenge/internal/challenge"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/logger"
)

const (
	jsonExt string = ".json"
)

func TestChallengesExpectPass(t *testing.T) {
	log := logger.NewLogger()
	log.SetOutput(ioutil.Discard)
	cfg := config.New()
	cfg.RunCmdTimeout = time.Second * 10
	chDir := path.Join(cfg.ROVolumeDir, "ch")

	items, _ := ioutil.ReadDir(chDir)
	for _, item := range items {
		if filepath.Ext(item.Name()) != jsonExt {
			continue
		}
		ch, err := challenge.New(path.Join(chDir, item.Name()))
		if err != nil {
			t.Fatalf("Unable to open challenge file `%s`: %s", item.Name(), err.Error())
		}
		t.Run(ch.Slug(), func(t *testing.T) {
			t.Parallel()
			runner := New(log, cfg)
			result, err := runner.RunContainer(ch, ch.Example())
			if err != nil {
				t.Fatalf("Unable to run container: %s", err.Error())
			}
			if *result.Correct != true {
				t.Fatalf("Got incorrect answer for %s", ch.Slug())
			}

			if result.Error != nil {
				t.Fatalf("Got test errors for %s: %v", ch.Slug(), *result.Error)
			}

			if result.TestPass != nil && *result.TestPass == false {
				t.Fatalf("Tests didn't pass for %s", ch.Slug())
			}

			if result.AfterRandOutputPass != nil && *result.AfterRandOutputPass == false {
				t.Fatalf("Output after random data didn't pass for %s", ch.Slug())
			}

			if result.AfterRandTestPass != nil && *result.AfterRandTestPass == false {
				t.Fatalf("Tests after random data didn't pass for %s", ch.Slug())
			}
		})
	}
}

func TestChallengesExpectFail(t *testing.T) {
	log := logger.NewLogger()
	log.SetOutput(ioutil.Discard)
	cfg := config.New()
	cfg.RunCmdTimeout = time.Second * 10
	chDir := path.Join(cfg.ROVolumeDir, "ch")

	items, _ := ioutil.ReadDir(chDir)
	for _, item := range items {
		if filepath.Ext(item.Name()) != jsonExt {
			continue
		}
		ch, err := challenge.New(path.Join(chDir, item.Name()))
		if err != nil {
			t.Fatalf("Unable to open challenge file `%s`: %s", item.Name(), err.Error())
		}
		t.Run(ch.Slug(), func(t *testing.T) {
			t.Parallel()
			for _, failure := range ch.ExpectedFailures() {
				runner := New(log, cfg)
				result, err := runner.RunContainer(ch, failure)
				if err != nil {
					t.Fatalf("Unable to run container: %s", err.Error())
				}

				if result.Correct == nil {
					t.Fatalf("Received an invalid challenge response, errors: %v", result.Error)
				}

				if *result.Correct == true {
					t.Fatalf("Got correct answer for %s", ch.Slug())
				}

				if result.TestPass != nil {
					if *result.TestPass == true {
						t.Fatalf("Tests passed with expected failure for %s", ch.Slug())
					}
					if result.Error == nil {
						t.Fatalf("Tests didn't pass but didn't get test errors for %s", ch.Slug())
					}
				}

				if result.AfterRandTestPass != nil {
					if *result.AfterRandTestPass == true {
						t.Fatalf("After random data, tests passed with expected failure for %s", ch.Slug())
					}
					if result.Error == nil {
						t.Fatalf("After random data, rests didn't pass but didn't get test errors for %s", ch.Slug())
					}
				}
			}
		})
	}
}
