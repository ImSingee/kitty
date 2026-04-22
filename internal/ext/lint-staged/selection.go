package lintstaged

import "fmt"

type SelectionMode string

const (
	SelectionModeStaged    SelectionMode = "staged"
	SelectionModeUnstaged  SelectionMode = "unstaged"
	SelectionModeUntracked SelectionMode = "untracked"
	SelectionModeTracked   SelectionMode = "tracked"
	SelectionModeChanged   SelectionMode = "changed"
	SelectionModeAll       SelectionMode = "all"
)

func (o *Options) SelectionMode() SelectionMode {
	if o.Status == "" {
		return SelectionModeStaged
	}

	return SelectionMode(o.Status)
}

func (o *Options) UsesIndex() bool {
	return o.Diff == "" && o.SelectionMode() == SelectionModeStaged
}

func (o *Options) ValidateSelectionMode() error {
	switch o.SelectionMode() {
	case SelectionModeStaged, SelectionModeUnstaged, SelectionModeUntracked, SelectionModeTracked, SelectionModeChanged, SelectionModeAll:
		// ok
	default:
		return fmt.Errorf("invalid --status %q (must be one of: staged, unstaged, untracked, tracked, changed, all)", o.Status)
	}

	if o.Diff != "" && o.SelectionMode() != SelectionModeStaged {
		return fmt.Errorf("--diff cannot be used together with --status=%s", o.SelectionMode())
	}

	return nil
}

func (o *Options) SelectionReason() string {
	if o.Diff != "" {
		return "`--diff` was used"
	}
	if o.SelectionMode() != SelectionModeStaged {
		return fmt.Sprintf("`--status=%s` was used", o.SelectionMode())
	}

	return ""
}

func (o *Options) SelectedFilesLabel() string {
	if o.Diff != "" {
		return "selected files"
	}

	switch o.SelectionMode() {
	case SelectionModeUnstaged:
		return "unstaged files"
	case SelectionModeUntracked:
		return "untracked files"
	case SelectionModeTracked:
		return "tracked changed files"
	case SelectionModeChanged:
		return "changed files"
	case SelectionModeAll:
		return "all files"
	default:
		return "staged files"
	}
}
