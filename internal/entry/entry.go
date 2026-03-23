package entry

import "time"

type Kind string

const (
	KindCommit Kind = "commit"
	KindNote   Kind = "note"
)

type Entry struct {
	Kind          Kind
	CapturedAt    time.Time
	RepoName      string
	RepoPath      string
	Branch        string
	CommitSHA     string
	CommitMessage string
	FilesChanged  []string
	DiffStat      string
	Note          string
}
