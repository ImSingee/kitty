package tl

type TaskCallback interface {
	GetTask() *Task
	//Log(msg string)
	Hide()
	Skip(reason string)
	AddSubTaskList(taskList *TaskList)
	AddSubTask(tasks ...*Task)
}
