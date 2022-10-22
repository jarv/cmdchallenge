package store

type CmdStore struct {
	Cmd      *string
	Slug     *string
	Version  *int
	Correct  *bool
	Error    *string
	ExitCode *int
	Output   *string
}

type CmdStorer interface {
	GetResult(cmd, slug string, version int) (*CmdStore, error)
	CreateResult(s *CmdStore) error
	IncrementResult(cmd, slug string, version int) error
	TopCmdsForSlug(slug string) ([]string, error)
}
