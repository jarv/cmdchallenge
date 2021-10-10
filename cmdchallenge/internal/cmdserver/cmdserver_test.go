package cmdserver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jarv/cmdchallenge/internal/challenge"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/logger"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"gitlab.com/jarv/cmdchallenge/internal/runner"
)

const ROVolumeDir string = "../../ro_volume"

var (
	correctResult = runner.RunnerResult{
		Output:              strPtr("fake output"),
		ExitCode:            intPtr(0),
		Correct:             boolPtr(true),
		OutputPass:          boolPtr(true),
		TestPass:            nil,
		AfterRandOutputPass: nil,
		Error:               nil,
	}

	helloWorldChallenge, _ = challenge.New(path.Join(ROVolumeDir, "ch", "hello_world.json"))
	cfg                    = &config.Config{ROVolumeDir: ROVolumeDir}
	log                    = logger.NewLogger()
	m                      = metrics.New()
)

type StubResultStor struct {
	mock.Mock
}

func (c *StubResultStor) GetResult(fingerprint string) (*runner.RunnerResult, error) {
	args := c.Called(fingerprint)

	return args.Get(0).(*runner.RunnerResult), args.Error(1)
}

func (c *StubResultStor) CreateResult(fingerprint, cmd, slug string, version int, result *runner.RunnerResult) error {
	c.Called(fingerprint, result)
	return nil
}

func (c *StubResultStor) IncrementResult(fingerprint string) error {
	c.Called(fingerprint)
	return nil
}

func (c *StubResultStor) TopCmdsForSlug(slug string) ([]string, error) {
	return make([]string, 0), nil
}

type StubRunnerExecutor struct {
	mock.Mock
}

func (s *StubRunnerExecutor) PullImages() error {
	args := s.Called()

	return args.Error(0)
}

func (s *StubRunnerExecutor) RunContainer(ch *challenge.Challenge, cmd string) (*runner.RunnerResult, error) {
	args := s.Called(ch, cmd)

	return args.Get(0).(*runner.RunnerResult), args.Error(1)
}

func TestRequest(t *testing.T) {
	req, resp := createTestRequest()
	log.SetOutput(ioutil.Discard)

	// Expectations for Result Stor
	stubResultStor := &StubResultStor{}

	stubResultStor.On(
		"GetResult",
		"370ebb46424ce444b722acd2783b03475f99a0c3e2bde5ad8e6d6b619df81c35",
	).Return(&correctResult, runner.ErrResultNotFound).Once()

	stubResultStor.On(
		"CreateResult",
		"370ebb46424ce444b722acd2783b03475f99a0c3e2bde5ad8e6d6b619df81c35",
		&correctResult,
	).Once()

	stubResultStor.On(
		"IncrementResult",
		"370ebb46424ce444b722acd2783b03475f99a0c3e2bde5ad8e6d6b619df81c35",
	).Once()

	// Expectation for Runner Executor
	stubRunnerExecutor := &StubRunnerExecutor{}

	stubRunnerExecutor.On(
		"RunContainer",
		helloWorldChallenge,
		"echo hello world",
	).Return(&correctResult, nil)

	s := New(log, cfg, m, stubRunnerExecutor, stubResultStor, false)
	s.runHandler(resp, req)

	stubResultStor.AssertExpectations(t)
	stubRunnerExecutor.AssertExpectations(t)

	jsonResp := resp.Body.String()
	expectedResp := `{"Cached":false,"Correct":true,"ExitCode":0,"Output":"fake output"}`
	assert.Equal(t, jsonResp, expectedResp)

	result := runner.RunnerResult{}

	assert.Nil(t, json.Unmarshal([]byte(jsonResp), &result))
	assert.Equal(t, resp.Code, 200)
}

func TestRequestCached(t *testing.T) {
	req, resp := createTestRequest()
	log.SetOutput(ioutil.Discard)

	// Expectations for Result Stor
	stubResultStor := &StubResultStor{}
	stubResultStor.On(
		"GetResult",
		"370ebb46424ce444b722acd2783b03475f99a0c3e2bde5ad8e6d6b619df81c35",
	).Return(&correctResult, nil).Once()

	stubResultStor.On(
		"IncrementResult",
		"370ebb46424ce444b722acd2783b03475f99a0c3e2bde5ad8e6d6b619df81c35",
	).Once()

	// Expectation for Runner Executor
	stubRunnerExecutor := &StubRunnerExecutor{}

	s := New(log, cfg, m, stubRunnerExecutor, stubResultStor, false)
	s.runHandler(resp, req)

	stubResultStor.AssertExpectations(t)
	stubRunnerExecutor.AssertExpectations(t)

	jsonResp := resp.Body.String()
	expectedResp := `{"Cached":true,"Correct":true,"ExitCode":0,"Output":"fake output"}`
	assert.Equal(t, jsonResp, expectedResp)
}

func createTestRequest() (*http.Request, *httptest.ResponseRecorder) {
	data := url.Values{}
	data.Set("cmd", "echo hello world")
	data.Set("slug", "hello_world")
	req, _ := http.NewRequest(http.MethodPost, "/c/r", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.1.1.1:80"

	return req, httptest.NewRecorder()
}
