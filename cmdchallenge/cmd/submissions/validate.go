package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"gitlab.com/jarv/cmdchallenge/internal/challenge"
	"gitlab.com/jarv/cmdchallenge/internal/config"
)

type Answer struct {
	Cmd     string `json:"cmd"`
	Slug    string `json:"slug"`
	Correct int    `json:"correct"`
}

func noError(err error) {
	if err != nil {
		panic(err)
	}
}

func newLogger(color bool) logr.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339, NoColor: !color}
	zl := zerolog.New(output).With().Caller().Timestamp().Logger()
	zerologr.VerbosityFieldName = ""

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	return zerologr.New(&zl)
}

func correct(c int) bool {
	return c == 1
}
func disp(results []string) {
	for _, r := range results {
		fmt.Printf("RESULTS\n-------------------------\n")
		fmt.Println(r)
	}
}

func main() {
	cfg := config.New(config.ConfigOpts{})
	log := newLogger(true)
	discardLog := logr.Discard()

	type Answer struct {
		Cmd     string `json:"cmd"`
		Slug    string `json:"slug"`
		Correct int    `json:"correct"`
	}

	f, err := os.Open("cmd/submissions/testdata/user-submissions.json.gz")
	noError(err)
	defer f.Close()

	gr, err := gzip.NewReader(f)
	noError(err)
	defer gr.Close()

	answers := []Answer{}
	decoder := json.NewDecoder(gr)
	noError(decoder.Decode(&answers))

	results := []string{}

	log.Info(fmt.Sprintf("Checking %d submissions", len(answers)))

	for i := 1850; i < len(answers); i++ {
		a := answers[i]
		if i%50 == 0 {
			if len(results) == 0 {
				log.Info(fmt.Sprintf("> %d", i))
			} else {
				log.Info(fmt.Sprintf("> %d --->%d<---", i, len(results)))
			}
		}
		ch, err := challenge.NewChallenge(challenge.ChallengeOptions{Slug: a.Slug})
		noError(err)

		runner := challenge.NewRunner(discardLog, cfg)
		result, err := runner.RunContainer(a.Cmd, ch)
		noError(err)
		if result.Correct == nil {
			results = append(results, fmt.Sprintf("got:nil, want:%v cmd:%s slug=%s", correct(a.Correct), a.Cmd, a.Slug))
			disp(results)
		}

		if correct(a.Correct) != *result.Correct {
			results = append(results, fmt.Sprintf("got %v, want:%v cmd:%s slug=%s", *result.Correct, correct(a.Correct), a.Cmd, a.Slug))
			disp(results)
		}
	}
	disp(results)
}
