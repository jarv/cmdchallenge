package challenge

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"gitlab.com/jarv/cmdchallenge/internal/store"
)

var (
	fakeResponse = CmdResponse{
		Correct:  toPtr(true),
		ExitCode: toPtr(0),
		Output:   toPtr("hello world"),
	}
	fakeStore = store.CmdStore{
		Cmd:      toPtr("echo hello world"),
		Slug:     toPtr("hello_world"),
		Version:  toPtr(5),
		Correct:  toPtr(true),
		ExitCode: toPtr(0),
		Output:   toPtr("hello world"),
	}
)

var cfg = config.New(config.ConfigOpts{})

type StubStor struct {
	mock.Mock
}

func (c *StubStor) GetResult(cmd, slug string, version int) (*store.CmdStore, error) {
	args := c.Called(cmd, slug, version)

	if args[0] == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*store.CmdStore), args.Error(1)
}

func (c *StubStor) CreateResult(s *store.CmdStore) error {
	args := c.Called(s)
	return args.Error(0)
}

func (c *StubStor) IncrementResult(cmd, slug string, version int) error {
	args := c.Called(cmd, slug, version)
	return args.Error(0)
}

func (c *StubStor) TopCmdsForSlug(slug string) ([]string, error) {
	return make([]string, 0), nil
}

type StubRunnerExecutor struct {
	mock.Mock
}

func (s *StubRunnerExecutor) PullImages() error {
	args := s.Called()

	return args.Error(0)
}

func (s *StubRunnerExecutor) RunContainer(cmd string, ch *Challenge) (*CmdResponse, error) {
	args := s.Called(cmd, ch)

	return args.Get(0).(*CmdResponse), args.Error(1)
}

func TestRequestNoCache(t *testing.T) {
	req, resp := createTestRequest()
	stubStore := &StubStor{}

	stubStore.On(
		"GetResult",
		"echo hello world",
		"hello_world",
		5,
	).Return(nil, store.ErrResultNotFound).Once()

	stubStore.On(
		"CreateResult",
		&fakeStore,
	).Once().Return(nil)

	stubStore.On(
		"IncrementResult",
		"echo hello world",
		"hello_world",
		5,
	).Once().Return(nil)

	// Expectation for Runner Executor
	stubRunnerExecutor := &StubRunnerExecutor{}

	stubRunnerExecutor.On(
		"RunContainer",
		"echo hello world",
		helloWorldCh(t),
	).Return(&fakeResponse, nil)

	s := NewServer(testr.New(t), cfg, metrics.New(testr.New(t)), stubRunnerExecutor, stubStore)
	s.runHandler(resp, req)

	stubStore.AssertExpectations(t)
	stubRunnerExecutor.AssertExpectations(t)

	jsonResp := resp.Body.String()
	expectedResp := `{"Cached":false,"Correct":true,"ExitCode":0,"Output":"hello world"}`
	assert.Equal(t, expectedResp, jsonResp)

	result := CmdResponse{}
	assert.Nil(t, json.Unmarshal([]byte(jsonResp), &result))

	assert.Equal(t, 200, resp.Code)
}

func TestRequestCached(t *testing.T) {
	req, resp := createTestRequest()

	// Expectations for Result Stor
	stubStore := &StubStor{}

	stubStore.On(
		"GetResult",
		"echo hello world",
		"hello_world",
		5,
	).Return(&fakeStore, nil).Once()

	stubStore.On(
		"IncrementResult",
		"echo hello world",
		"hello_world",
		5,
	).Once().Return(nil)

	// Expectation for Runner Executor
	stubRunnerExecutor := &StubRunnerExecutor{}

	s := NewServer(testr.New(t), cfg, metrics.New(testr.New(t)), stubRunnerExecutor, stubStore)
	s.runHandler(resp, req)

	stubStore.AssertExpectations(t)
	stubRunnerExecutor.AssertExpectations(t)

	jsonResp := resp.Body.String()
	expectedResp := `{"Cached":true,"Correct":true,"ExitCode":0,"Output":"hello world"}`
	assert.Equal(t, expectedResp, jsonResp)

	result := CmdResponse{}
	assert.Nil(t, json.Unmarshal([]byte(jsonResp), &result))

	assert.Equal(t, 200, resp.Code)
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

func helloWorldCh(t *testing.T) *Challenge {
	ch, err := NewChallenge(ChallengeOptions{Slug: "hello_world"})
	require.NoError(t, err)
	return ch
}
