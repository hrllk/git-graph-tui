package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"hrllk/graphkeeper/internal/git"
	"hrllk/graphkeeper/internal/state"
)

func (m model) handleBlockedKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.status = dismissAlertStatus(m.repoStatus)
		return m, nil
	default:
		return m, nil
	}
}

func dismissAlertStatus(repoStatus git.Status) state.Status {
	status := state.New().WithBrowse()
	if repoStatus.WorktreeDirty {
		status.WorktreeState = state.WorktreeStateDirty
	} else {
		status.WorktreeState = state.WorktreeStateClean
	}
	return status
}
