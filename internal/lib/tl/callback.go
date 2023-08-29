package tl

type TaskCallback interface {
	//Log(msg string)
	Hide()
	Skip(reason string)
	AddSubTaskList(taskList *TaskList)
	AddSubTask(tasks ...*Task)
}
