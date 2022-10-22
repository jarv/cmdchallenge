package challenge

import "errors"

var (
	ErrRunnerNonZeroReturn     = errors.New("non-zero return code")
	ErrRunnerDecodeResult      = errors.New("unable to decode result")
	ErrRunnerResultNotFound    = errors.New("result not found")
	ErrRunnerTimeout           = errors.New("runner timeout")
	ErrRunnerImgRemovalTimeout = errors.New("unable to cleanup after timeout")
)

type ChallengeError struct {
	msg string
	typ string
}

func (s *ChallengeError) Error() string {
	return s.msg
}

var (
	ErrServerInvalidSourceIP  = errors.New("unable to determine source IP")
	ErrServerCmdTooLong       = errors.New("command is too long")
	ErrServerInvalidMethod    = errors.New("invalid method")
	ErrServerInvalidRequest   = errors.New("request must include slug and cmd")
	ErrServerInvalidChallenge = errors.New("invalid challenge")
	ErrServerUnknown          = errors.New("unknown error")
	ErrServerDecode           = errors.New("decode error")
)

const (
	RunnerError     = "runner error"
	RunnerTimeout   = "timed out executing command"
	RunCmdInvalid   = "invalid response from runcmd"
	StoreError      = "storage error"
	StoreQueryError = "storage query error"
	TypeStore       = "store"
	TypeServer      = "server"
	TypeRunCmd      = "runcmd"
	TypeRunner      = "runner"
)

var (
	ErrSolutionsInvalidMethod = errors.New("invalid method for solutions")
	ErrSolutionsInvalidParam  = errors.New("invalid parameter for solutions")
	ErrSolutionsStore         = errors.New("storage error for solutions")
)

var (
	ErrCheckNotExist        = errors.New("check does not exist")
	ErrOopsProccessNeverRan = errors.New("the oops process was never ran")
)
