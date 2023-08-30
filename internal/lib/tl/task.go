package tl

import (
	"fmt"
	"github.com/ImSingee/go-ex/ee"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"strings"
)

type Task struct {
	Title string
	Run   func(callback TaskCallback) error
	// PostRun runs after Run (even if Run failed) with result
	PostRun func(result *Result)
	Enable  func() bool

	Options []OptionApplier

	id string
	option
}

func NewTask(title string, run func(callback TaskCallback) error, options ...OptionApplier) *Task {
	return &Task{
		Title:   title,
		Run:     run,
		Options: options,
	}
}

// Id only available after run
func (t *Task) Id() string {
	return t.id
}

func (t *Task) use() {
	if t.id != "" {
		panic("Cannot use the same task more than once")
	}

	t.id = uuid.NewString()
	if !t.inited {
		panic("Task can only be used inside TaskList (this is internal logic error)")
	}

	for _, applyOpt := range t.Options {
		applyOpt(&t.option)
	}
}

func (t *Task) start(p *tea.Program) (result *Result) {
	result = &Result{
		Task:    t,
		Enabled: t.shouldEnable(),
	}

	if !result.Enabled {
		return
	}

	p.Send(&eventTaskStart{
		Id: t.id,
	})

	// update UI
	defer func() {
		if result.Hide {
			p.Send(&eventTaskHide{
				Id: t.id,
			})
		}

		if result.Error {
			p.Send(&eventTaskFail{
				Id:  t.id,
				Err: result.Err,
			})
			return
		}

		if result.Skipped {
			p.Send(&eventTaskSkip{
				Id:     t.id,
				Reason: result.SkipReason,
			})
			return
		}

		p.Send(&eventTaskSuccess{
			Id: t.id,
		})
	}()

	controller := t.controller()

	defer func() {
		t.postRun(result)
	}()

	result.Err = t.run(controller)
	if result.Err != nil {
		result.Error = true
	}

	if controller.hide {
		result.Hide = true
	}

	if result.Err != nil {
		return
	}

	if controller.skipped {
		result.Skipped = true
		result.SkipReason = controller.skipReason
		return
	}

	if controller.subList != nil {
		controller.subList.option = t.option
		controller.subList.prepare()

		p.Send(&eventTaskAddSubList{
			Id:   t.id,
			List: controller.subList.createModel(),
		})

		subListResult := controller.subList.start(p)
		result.TaskList = subListResult.TaskList
		result.SubResults = subListResult.SubResults

		if subListResult.Error {
			result.Error = true // without reason
			return
		}
	}

	return
}

func (t *Task) run(controller *taskController) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
		}
	}()

	if t.Run == nil {
		return ee.New("no Run function provided")
	}

	return t.Run(controller)
}

func (t *Task) postRun(result *Result) {
	defer func() {
		if e := recover(); e != nil {
			if result.Err != nil {
				result.Err = fmt.Errorf("PostRun panic: %v; Run errors: %w", e, result.Err)
			} else if result.Error {
				result.Err = fmt.Errorf("PostRun panic: %v; Run contains anoynomous errors", e)
			} else {
				result.Err = fmt.Errorf("PostRun panic: %v", e)
			}

			result.Error = true
		}
	}()

	if t.PostRun != nil {
		t.PostRun(result)
	}
}

func (t *Task) controller() *taskController {
	return &taskController{
		task: t,
	}
}

func (t *Task) skip(p *tea.Program) {
	p.Send(&eventTaskSkip{
		Id: t.id,
	})
}

type taskModel struct {
	id          string
	title       string
	status      taskStatus
	skipReason  string
	enable      bool
	hide        bool
	errorReason string
	subList     tea.Model
}

func (t *Task) shouldEnable() bool {
	if t.Enable == nil {
		return true
	}
	return t.Enable()
}

type taskStatus uint8

const (
	taskStatusPending taskStatus = iota
	taskStatusRunning
	taskStatusSuccess
	taskStatusFailed
	taskStatusSkipped
)

func (t *Task) createModel() taskModel {
	m := taskModel{
		id:     t.id,
		title:  t.Title,
		status: taskStatusPending,
		enable: t.shouldEnable(),
		hide:   false,
	}

	return m
}

func (m taskModel) Init() tea.Cmd {
	return nil
}

// model msg cmd

func (m taskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case *eventTaskStart:
		if m.id == v.Id {
			m.enable = true
			m.status = taskStatusRunning
			return m, nil
		}
	case *eventTaskSuccess:
		if m.id == v.Id {
			m.status = taskStatusSuccess
			return m, nil
		}
	case *eventTaskFail:
		if m.id == v.Id {
			m.status = taskStatusFailed
			if v.Err != nil {
				m.errorReason = v.Err.Error()
			}
			return m, nil
		}
	case *eventTaskSkip:
		if m.id == v.Id {
			m.status = taskStatusSkipped
			m.skipReason = v.Reason
			return m, nil
		}
	case *eventTaskHide:
		if m.id == v.Id {
			m.hide = true
			return m, nil
		}
	case *eventTaskAddSubList:
		if m.id == v.Id {
			m.subList = v.List
			return m, nil
		}
	}

	if m.subList != nil {
		l, cmd := m.subList.Update(msg)
		m.subList = l
		return m, cmd
	}

	return m, nil
}

func (m taskModel) View() string {
	if !m.enable || m.hide {
		return ""
	}

	b := strings.Builder{}

	icon := "○"
	switch m.status {
	case taskStatusPending:
		icon = "○"
	case taskStatusRunning:
		icon = ">"
	case taskStatusSuccess:
		icon = "✓"
	case taskStatusFailed:
		icon = "✗"
	case taskStatusSkipped:
		icon = "-"
	}
	b.WriteString(icon + " ")

	b.WriteString(m.title)

	if m.status == taskStatusSkipped {
		b.WriteString(" (skipped")
		if m.skipReason != "" {
			b.WriteString(" - ")
			b.WriteString(m.skipReason)
		}
		b.WriteString(")")
	}

	b.WriteString("\n")

	if m.errorReason != "" {
		b.WriteString("  ERROR: ")
		b.WriteString(strings.TrimSpace(m.errorReason))
		b.WriteString("\n")
	}

	if m.subList != nil {
		subListView := m.subList.View()
		subListView = strings.TrimSpace(subListView)

		if len(subListView) != 0 {
			for _, line := range strings.Split(subListView, "\n") {
				b.WriteString("  ")
				b.WriteString(line)
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

type taskController struct {
	task *Task

	skipped    bool
	skipReason string
	hide       bool
	subList    *TaskList
}

func (c *taskController) GetTask() *Task {
	return c.task
}

//func (c *taskController) Log(msg string) {
//
//}

func (c *taskController) Skip(reason string) {
	c.skipped = true
	c.skipReason = reason
}

func (c *taskController) Hide() {
	c.hide = true
}

func (c *taskController) AddSubTaskList(taskList *TaskList) {
	if c.subList != nil {
		panic("One task only support one subList now")
	}

	c.subList = taskList
}

func (c *taskController) AddSubTask(tasks ...*Task) {
	if c.subList == nil {
		c.subList = NewTaskList(nil)
	}

	c.subList.Tasks = append(c.subList.Tasks, tasks...)
}
