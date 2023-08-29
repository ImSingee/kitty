package tl

import (
	"github.com/ImSingee/go-ex/mr"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"strings"
)

type TaskList struct {
	Tasks   []*Task
	Options []OptionApplier

	id    string
	tasks []*Task
	option
}

func NewTaskList(tasks []*Task, options ...OptionApplier) *TaskList {
	return &TaskList{
		Tasks:   tasks,
		Options: options,
	}
}

func (tl *TaskList) use() {
	if tl.id != "" {
		panic("Cannot use the same task list more than once")
	}

	tl.id = uuid.NewString()

	if !tl.inited {
		tl.option = defaultOption()
	}

	for _, applyOpt := range tl.Options {
		applyOpt(&tl.option)
	}

	tl.tasks = tl.Tasks
	for _, task := range tl.tasks {
		task.option = tl.option
		task.use()
	}
}

type tlModel struct {
	id    string
	tasks []tea.Model
}

func (tl *TaskList) createModel() tlModel {
	return tlModel{
		id: tl.id,
		tasks: mr.Map(tl.tasks, func(t *Task, _index int) tea.Model {
			return t.createModel()
		}),
	}
}

func (m tlModel) Init() tea.Cmd {
	return nil
}

func (m tlModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, len(m.tasks))

	for i := range m.tasks {
		t, cmd := m.tasks[i].Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.tasks[i] = t
	}

	return m, tea.Batch(cmds...)
}

func (m tlModel) View() string {
	views := mr.Map(m.tasks, func(task tea.Model, index int) string {
		return task.View()
	})

	return strings.Join(views, "")
}

func (tl *TaskList) prepare() {
	tl.use()
}

func (tl *TaskList) start(p *tea.Program) (iAmError bool) {
	preventContinue := false

	for _, task := range tl.tasks {
		if preventContinue {
			task.skip(p)
			continue
		}

		taskIsError := task.start(p)
		if taskIsError {
			iAmError = true
			if tl.exitOnError && task.exitOnError {
				preventContinue = true // stop current task list execution
			}
		}
	}
	return
}