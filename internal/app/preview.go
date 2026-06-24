package app

import (
	"fmt"

	"hrllk/graphkeeper/internal/git"
	"hrllk/graphkeeper/internal/state"
)

func buildActionPreview(action state.Action, target string, rs git.Status, currentOnly, targetOnly int) state.Status {
	switch action {
	case state.ActionMerge:
		return buildMergePreview(target, rs, currentOnly, targetOnly)
	case state.ActionRebase:
		return buildRebasePreview(target, currentOnly, targetOnly)
	case state.ActionReset:
		return buildResetPreview(target, rs, currentOnly, targetOnly)
	default:
		return state.New().WithOutcome(action, "No action selected.", target, false)
	}
}

func buildMergePreview(target string, rs git.Status, currentOnly, targetOnly int) state.Status {
	switch {
	case currentOnly == 0 && targetOnly == 0:
		return state.New().WithOutcome(state.ActionMerge, "Target already matches HEAD.", "Nothing moves. The branch already points at the same commit.", true)
	case currentOnly == 0:
		return state.New().WithOutcome(state.ActionMerge, "FF 가능. 포인터만 이동합니다.", "HEAD can move to "+target+". "+countDetail(currentOnly, targetOnly), true)
	case targetOnly == 0:
		return state.New().WithOutcome(state.ActionMerge, "대상은 이미 포함되어 있습니다.", "Current branch already contains "+target+". "+countDetail(currentOnly, targetOnly), true)
	default:
		return state.New().WithOutcome(state.ActionMerge, "FF 불가. merge commit이 필요합니다.", "HEAD "+shorten(rs.Head, 12)+" and target "+target+" have diverged. "+countDetail(currentOnly, targetOnly), true)
	}
}

func buildRebasePreview(target string, currentOnly, targetOnly int) state.Status {
	switch {
	case currentOnly == 0 && targetOnly == 0:
		return state.New().WithOutcome(state.ActionRebase, "Target already matches HEAD.", "Nothing is rewritten because both refs point at the same commit.", true)
	case targetOnly == 0:
		return state.New().WithOutcome(state.ActionRebase, "Target is already in the base history.", "Current commits will replay onto "+target+". "+countDetail(currentOnly, targetOnly), true)
	default:
		return state.New().WithOutcome(state.ActionRebase, "새 base 위로 커밋을 재배치합니다.", countDetail(currentOnly, targetOnly)+"  |  target: "+target, true)
	}
}

func buildResetPreview(target string, rs git.Status, currentOnly, targetOnly int) state.Status {
	return state.New().WithOutcome(state.ActionReset, "현재 HEAD를 선택한 위치로 이동합니다.", "HEAD "+shorten(rs.Head, 12)+" -> "+target+"  |  "+countDetail(currentOnly, targetOnly), true)
}

func countDetail(currentOnly, targetOnly int) string {
	return "Current-only: " + fmt.Sprint(currentOnly) + "  Target-only: " + fmt.Sprint(targetOnly)
}
