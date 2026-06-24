package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"hrllk/graphkeeper/internal/graph"
	"hrllk/graphkeeper/internal/state"
	"hrllk/graphkeeper/internal/telemetry"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case loadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = m.status.WithError(msg.err.Error())
			telemetry.Log("app", "load_error", map[string]string{"error": msg.err.Error()})
			return m, nil
		}
		m.repoStatus = msg.status
		syncBrowseState(&m, msg.status)
		m.status = deriveStatus(msg.status)
		telemetry.Log("app", "load_repo", map[string]string{
			"root":   msg.status.Root,
			"branch": msg.status.Branch,
			"head":   msg.status.Head,
		})
		return m, nil
	case tickMsg:
		return m, tea.Batch(scheduleRefresh(), refreshRepoState(m.repo, m.commitLimit))
	case refreshedMsg:
		if msg.err != nil {
			return m, nil
		}
		m.repoStatus = msg.status
		syncBrowseState(&m, msg.status)
		if !m.branchOpen && (m.status.Mode == state.ModeBrowse || m.status.Mode == state.ModeBlocked || m.status.Mode == state.ModeEmpty || m.status.Mode == state.ModeError) {
			m.status = deriveStatus(msg.status)
		}
		return m, nil
	case fetchedMsg:
		if msg.err != nil {
			m.status = state.New().WithBlocked(state.BlockFetchFailed, "Fetch failed.", msg.err.Error())
			return m, nil
		}
		m.repoStatus = msg.status
		syncBrowseState(&m, msg.status)
		if m.status.Mode == state.ModeBrowse || m.status.Mode == state.ModeBlocked || m.status.Mode == state.ModeEmpty || m.status.Mode == state.ModeError {
			m.status = deriveStatus(msg.status)
		}
		telemetry.Log("app", "fetch_repo", map[string]string{
			"branch": msg.status.Branch,
			"head":   msg.status.Head,
		})
		return m, nil
	case preparedMsg:
		if msg.err != nil {
			m.status = state.New().WithBlocked(state.BlockFetchFailed, "Fetch failed.", msg.err.Error())
			telemetry.Log("app", "prepare_failed", map[string]string{"action": string(msg.action), "error": msg.err.Error()})
			return m, nil
		}
		m.repoStatus = msg.status
		syncBrowseState(&m, msg.status)
		switch msg.action {
		case state.ActionMerge, state.ActionRebase, state.ActionReset:
			m.status = actionPickTargets(msg.status, msg.action)
		default:
			m.status = deriveStatus(msg.status)
		}
		telemetry.Log("app", "prepare_action", map[string]string{
			"action": string(msg.action),
			"branch": msg.status.Branch,
		})
		return m, nil
	case pullCheckedMsg:
		if msg.err != nil {
			m.status = state.New().WithBlocked(state.BlockFetchFailed, "Fetch failed.", msg.err.Error())
			telemetry.Log("app", "pull_check_failed", map[string]string{"error": msg.err.Error()})
			return m, nil
		}
		m.repoStatus = msg.repo
		syncBrowseState(&m, msg.repo)
		m.status = msg.status
		telemetry.Log("app", "pull_check", map[string]string{
			"upstream": msg.repo.Upstream,
			"blocked":  string(msg.status.Block),
		})
		return m, nil
	case previewMsg:
		if msg.err != nil {
			m.status = state.New().WithBlocked(state.BlockUnknown, "Preview failed.", msg.err.Error())
			telemetry.Log("app", "preview_failed", map[string]string{"action": string(msg.action), "target": msg.target, "error": msg.err.Error()})
			return m, nil
		}
		m.repoStatus = msg.repo
		syncBrowseState(&m, msg.repo)
		m.status = msg.status
		m.status.Selected = msg.target
		telemetry.Log("app", "preview_action", map[string]string{
			"action": string(msg.action),
			"target": msg.target,
			"mode":   string(msg.status.Mode),
		})
		return m, nil
	case pushFetchedMsg:
		if msg.err != nil {
			m.status = state.New().WithBlocked(state.BlockFetchFailed, "Fetch failed before push.", msg.err.Error())
			return m, nil
		}
		m.repoStatus = msg.status
		syncBrowseState(&m, msg.status)
		if msg.status.NoUpstream {
			branchName := msg.status.Branch
			titleMsg := "Push and Track Remote?"
			detailMsg := fmt.Sprintf("There is no upstream configured for the current branch. Do you want to push and set upstream tracking to origin/%s?", branchName)
			m.status = m.status.WithConfirm(state.ActionSetUpstream, titleMsg, detailMsg)
			m.status.Title = titleMsg
			return m, nil
		}
		m.status = state.New().WithLoading("Pushing commits...")
		return m, executePush(m.repo, msg.status.Branch, m.commitLimit)
	case pullFetchedMsg:
		if msg.err != nil {
			m.status = state.New().WithBlocked(state.BlockFetchFailed, "Fetch failed before pull.", msg.err.Error())
			return m, nil
		}
		m.repoStatus = msg.status
		syncBrowseState(&m, msg.status)
		track := m.repoStatus.Tracking[m.repoStatus.Branch]
		isFF := track.Behind > 0 && track.Ahead == 0
		m.status = state.New().WithLoading("Analyzing pull changes...")
		return m, loadPullPreviewCommits(m.repo, isFF)
	case pullPreviewReadyMsg:
		if msg.err != nil {
			m.status = state.New().WithBlocked(state.BlockUnknown, "Analysis failed.", msg.err.Error())
			return m, nil
		}
		m.handshakeCommits = make(map[string]bool)
		if msg.isFF {
			if len(msg.commits) > 0 {
				m.handshakeCommits[msg.commits[0]] = true
			}
		} else {
			for _, hash := range msg.commits {
				m.handshakeCommits[hash] = true
			}
		}
		m.pullIsFastForward = msg.isFF
		var titleMsg, detailMsg string
		if msg.isFF {
			titleMsg = "Do you want to continue?"
			detailMsg = "The branch will fast-forward to the highlighted target commit."
			m.status = m.status.WithConfirm(state.ActionPull, titleMsg, detailMsg)
		} else {
			titleMsg = "Choose Pull Integration"
			detailMsg = "The branches have diverged. Choose integration strategy:\n\nm: merge pull (recreates merge commit)\nr: rebase pull (replays commits)\nesc: cancel integration"
			m.status = m.status.WithConfirm(state.ActionPull, titleMsg, detailMsg)
		}
		m.status.Title = titleMsg
		return m, nil
	case executedMsg:
		if msg.err != nil {
			isAuthError := strings.Contains(msg.err.Error(), "Permission denied") ||
				strings.Contains(msg.err.Error(), "Authentication failed") ||
				strings.Contains(msg.err.Error(), "Could not read from remote repository")

			if msg.action == state.ActionPush && !isAuthError && (strings.Contains(msg.err.Error(), "[rejected]") || strings.Contains(msg.err.Error(), "non-fast-forward")) {
				status := m.repoStatus
				if msg.status.Root != "" {
					status = msg.status
				}
				m.repoStatus = status
				m.handshakeCommits = make(map[string]bool)
				if status.Head != "" {
					m.handshakeCommits[status.Head] = true
				}
				remoteHash := findRemoteCommitHash(status, status.Upstream)
				if remoteHash != "" {
					m.handshakeCommits[remoteHash] = true
				}
				branchName := status.Branch
				titleMsg := fmt.Sprintf("Force push to origin/%s?", branchName)
				detailMsg := fmt.Sprintf("The remote branch has different history. Force pushing will overwrite origin/%s history with your local commits. Continue?", branchName)
				m.status = m.status.WithConfirm(state.ActionForcePush, titleMsg, detailMsg)
				m.status.Title = titleMsg
				return m, nil
			}
			if (msg.action == state.ActionPull || msg.action == state.ActionPullMerge || msg.action == state.ActionPullRebase) && (msg.status.MergeInProgress || msg.status.RebaseInProgress) {
				m.repoStatus = msg.status
				syncBrowseState(&m, msg.status)
				m.status = state.New().WithBrowse()
				m.status.Message = "Pull stopped with conflicts."
				m.status.Detail = "Press enter to abort the in-progress merge/rebase."
				telemetry.Log("app", "execute_conflicted", map[string]string{
					"action": string(msg.action),
					"head":   msg.status.Head,
				})
				return m, nil
			}
			if msg.action == state.ActionMerge && msg.status.MergeInProgress {
				m.repoStatus = msg.status
				syncBrowseState(&m, msg.status)
				m.status = state.New().WithBrowse()
				m.status.Message = "Merge stopped with conflicts."
				m.status.Detail = "Resolve conflicts, then press 'a' to abort or commit to complete."
				telemetry.Log("app", "execute_conflicted", map[string]string{
					"action": string(msg.action),
					"head":   msg.status.Head,
				})
				return m, nil
			}
			if msg.action == state.ActionRebase && msg.status.RebaseInProgress {
				m.repoStatus = msg.status
				syncBrowseState(&m, msg.status)
				m.status = state.New().WithBrowse()
				m.status.Message = "Rebase stopped with conflicts."
				m.status.Detail = "Resolve conflicts, then press 'a' to abort or continue rebase."
				telemetry.Log("app", "execute_conflicted", map[string]string{
					"action": string(msg.action),
					"head":   msg.status.Head,
				})
				return m, nil
			}
			reason := state.BlockUnknown
			message := "Execution failed."
			detail := msg.err.Error()
			if msg.action == state.ActionCheckout {
				message = "Checkout failed."
				if strings.Contains(detail, "local changes") || strings.Contains(detail, "overwritten by checkout") {
					reason = state.BlockDirtyTree
					message = "Checkout blocked by local changes."
					detail = "Your local changes would be overwritten by checkout. Commit or stash them before switching."
				}
			} else if isAuthError && (msg.action == state.ActionPush || msg.action == state.ActionForcePush || msg.action == state.ActionSetUpstream) {
				message = "Authentication or Permission error."
				detail = "Please check your remote credentials or network connection: " + msg.err.Error()
			} else if msg.action == state.ActionPush || msg.action == state.ActionForcePush || msg.action == state.ActionSetUpstream {
				message = "Push failed."
			}
			m.status = state.New().WithBlocked(reason, message, detail)
			telemetry.Log("app", "execute_failed", map[string]string{"action": string(msg.action), "target": msg.target, "error": msg.err.Error()})
			return m, nil
		}
		m.repoStatus = msg.status
		if msg.action == state.ActionPush || msg.action == state.ActionForcePush || msg.action == state.ActionSetUpstream || msg.action == state.ActionPullMerge || msg.action == state.ActionPullRebase {
			m.handshakeCommits = make(map[string]bool)
			syncBrowseState(&m, msg.status)
			m.status = deriveStatus(msg.status)
			if msg.action == state.ActionPullMerge || msg.action == state.ActionPullRebase {
				m.status.Message = "Pull completed successfully."
			} else {
				m.status.Message = fmt.Sprintf("Push completed for %s.", msg.target)
			}
			telemetry.Log("app", "execute_action", map[string]string{
				"action": string(msg.action),
				"head":   msg.status.Head,
			})
			return m, nil
		}
		if msg.action == state.ActionCheckout {
			m.commitLimit = initialGraphCommitLimit
			rows := graph.Rows(msg.status)
			if len(rows) > 0 {
				m.sectionCursor[sectionGraph] = 0
				m.graphScroll = 0
				m.graphLaneCursor = graph.PointerLane(rows[0])
			}
			syncBrowseState(&m, msg.status)
			m.status = deriveStatus(msg.status)
			telemetry.Log("app", "execute_action", map[string]string{
				"action": string(msg.action),
				"target": msg.target,
				"head":   msg.status.Head,
			})
			return m, nil
		}
		if msg.action == state.ActionPull {
			syncBrowseState(&m, msg.status)
			m.status = deriveStatus(msg.status)
			telemetry.Log("app", "execute_action", map[string]string{
				"action": string(msg.action),
				"head":   msg.status.Head,
			})
			return m, nil
		}
		if msg.action == state.ActionAbort {
			m.handshakeCommits = make(map[string]bool)
			syncBrowseState(&m, msg.status)
			m.status = deriveStatus(msg.status)
			telemetry.Log("app", "execute_action", map[string]string{
				"action": string(msg.action),
				"head":   msg.status.Head,
			})
			return m, nil
		}
		if msg.action == state.ActionReset {
			rows := graph.Rows(msg.status)
			rowIdx := graph.FindRowByHash(rows, msg.status.Head)
			if rowIdx >= 0 {
				m.sectionCursor[sectionGraph] = rowIdx
				m.graphScroll = clampScroll(rowIdx, len(rows), graphPageSize(&m))
			}
			syncBrowseState(&m, msg.status)
			m.status = deriveStatus(msg.status)
			m.status.Message = fmt.Sprintf("Hard reset completed to %s.", shorten(msg.target, 7))
			telemetry.Log("app", "execute_action", map[string]string{
				"action": string(msg.action),
				"target": msg.target,
				"head":   msg.status.Head,
			})
			return m, nil
		}
		if msg.action == state.ActionMerge || msg.action == state.ActionRebase {
			rows := graph.Rows(msg.status)
			rowIdx := graph.FindRowByHash(rows, msg.status.Head)
			if rowIdx >= 0 {
				m.sectionCursor[sectionGraph] = rowIdx
				m.graphScroll = clampScroll(rowIdx, len(rows), graphPageSize(&m))
			}
		}
		syncBrowseState(&m, msg.status)
		m.status = state.New().WithOutcome(msg.action, "Completed.", executionDetail(msg.action, msg.target, msg.status), false)
		m.status.Selected = msg.target
		telemetry.Log("app", "execute_action", map[string]string{
			"action": string(msg.action),
			"target": msg.target,
			"head":   msg.status.Head,
		})
		return m, nil
	case createdBranchMsg:
		if msg.err != nil {
			m.branchOpen = false
			reason := state.BlockUnknown
			message := "Branch creation failed."
			detail := msg.err.Error()
			if strings.Contains(msg.err.Error(), "working tree is not clean") {
				reason = state.BlockDirtyTree
				message = "Working tree is not clean."
				detail = "Commit or stash local changes before creating and checking out a new branch."
			}
			m.status = state.New().WithBlocked(reason, message, detail)
			telemetry.Log("app", "branch_create_failed", map[string]string{"name": msg.name, "base": msg.base, "error": msg.err.Error()})
			return m, nil
		}
		m.branchOpen = false
		m.repoStatus = msg.status
		syncBrowseState(&m, msg.status)
		m.status = deriveStatus(msg.status)
		telemetry.Log("app", "branch_create", map[string]string{"name": msg.name, "base": msg.base})
		return m, nil
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}
	return m, nil
}
