package tl

type Result struct {
	Task     *Task
	TaskList *TaskList

	Enabled    bool
	Skipped    bool
	SkipReason string
	Hide       bool
	Error      bool
	Err        error

	SubResults []*Result
}
