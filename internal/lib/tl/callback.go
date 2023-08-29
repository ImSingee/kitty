package tl

type TaskCallback interface {
	//Log(msg string)
	Skip(reason string)
	AddSubTaskList(taskList *TaskList)
	AddSubTask(tasks ...*Task)
}
