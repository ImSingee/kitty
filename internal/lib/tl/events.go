package tl

type eventTaskStart struct {
	Id string
}

type eventTaskSuccess struct {
	Id string
}

type eventTaskFail struct {
	Id  string
	Err error
}

type eventTaskSkip struct {
	Id     string
	Reason string
}

type eventTaskAddSubList struct {
	Id   string
	List tlModel
}
