package helpers

type Job struct {
	ID       string
	Status   JobStatus
	OutPath  string
	Progress Progress
}

type JobStatus string

const (
	JobPending JobStatus = "pending"
	JobRunning JobStatus = "running"
	JobDone    JobStatus = "done"
	JobFailed  JobStatus = "failed"
)

type Progress struct {
	ProgressStatus ProgressStatus `json:"status"`
	Completion     float32        `json:"completion"`
}
type ProgressStatus string

const (
	Initializing ProgressStatus = "initializing"
	Extracting   ProgressStatus = "extracting"
	Compiling    ProgressStatus = "compiling"
)
