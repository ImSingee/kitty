package tl

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type Runner struct {
	tl *TaskList

	err  error
	done bool
}

func New(tasks []*Task, options ...OptionApplier) *Runner {
	return &Runner{
		tl: NewTaskList(tasks, options...),
	}
}

func (runner *Runner) Run() error {
	runner.prepare()
	p := tea.NewProgram(runner.createModel())

	go func() {
		runner.start(p)
	}()

	_, err := p.Run()
	if err != nil {
		return err
	}

	if runner.err != nil {
		return runner.err
	}
	if !runner.done {
		return fmt.Errorf("canceled")
	}

	return err
}

func (runner *Runner) prepare() {
	runner.tl.prepare()
}

func (runner *Runner) start(p *tea.Program) {
	defer func() {
		runner.done = true
		p.Send(tea.Quit())
	}()

	someTasksError := runner.tl.start(p)
	if someTasksError {
		runner.err = fmt.Errorf("some tasks error")
	}
}

type runnerModel struct {
	tl tea.Model
}

func (runner *Runner) createModel() runnerModel {
	return runnerModel{
		tl: runner.tl.createModel(),
	}
}

func (m runnerModel) Init() tea.Cmd {
	return nil
}
func (m runnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if shouldQuit(msg) {
		return m, tea.Quit
	}

	tl, cmd := m.tl.Update(msg)
	m.tl = tl
	return m, cmd
}
func (m runnerModel) View() string {
	return m.tl.View()
}

func shouldQuit(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return true
		}
	}

	return false
}
