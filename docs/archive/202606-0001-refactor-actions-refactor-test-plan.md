### Goal
`internal/app/actions.go`를 책임 단위로 나누고, 리팩토링 전후 동작을 테스트로 고정한다.

OOP 관점은 Java식 class hierarchy가 아니라 책임과 협력의 분리로 적용한다. `StatusPolicy`, `TargetBuilder`, `PreviewBuilder`, `DecorationParser` 역할을 작은 package-private 함수로 표현한다.

### Scope
- 코드 위치: `internal/app/actions.go`
- 테스트 위치: `internal/app/actions_test.go`
- 기본 원칙: 동작 변경 없이 먼저 테스트를 추가하고, 이후 helper를 추출한다.
- 새 package 생성은 하지 않는다.

### Review Notes
- `actions.go`만 쪼개는 것으로 끝내지 않는다. target 규칙은 `navigation.go`와 중복되므로, 공통 helper로 먼저 모은다.
- preview는 merge, rebase, reset 책임이 서로 다르므로, 함수 분리 후 파일 분리까지 열어둔다.
- execution 완료 메시지와 remote commit lookup은 `actions.go`에 붙여두기보다 별도 책임으로 분리한다.
- 공용 target helper는 `target_items.go` 같은 전용 파일로 모으는 쪽이 낫다.

### BEFORE
현재 `actions.go`는 여러 책임이 한 파일과 큰 함수 안에 섞여 있다.

```go
func actionPickTargets(rs git.Status, action state.Action) state.Status {
	if (action == state.ActionMerge || action == state.ActionRebase) && rs.Detached {
		return state.New().WithBlocked(
			state.BlockDetached,
			"Detached HEAD.",
			"Choose a branch before merging or rebasing.",
		)
	}

	targets := make([]state.TargetItem, 0, len(rs.LocalBranches)+len(rs.RemoteBranches)+len(rs.Tags))
	for _, name := range rs.LocalBranches {
		upstream, known := branchUpstream(rs, name)
		targets = append(targets, state.TargetItem{
			Kind:       state.TargetKindLocal,
			Name:       name,
			Ref:        name,
			NoUpstream: known && upstream == "",
		})
	}
	for _, name := range rs.RemoteBranches {
		if strings.HasSuffix(name, "/HEAD") {
			continue
		}
		targets = append(targets, state.TargetItem{
			Kind: state.TargetKindRemote,
			Name: name,
			Ref:  name,
		})
	}
	for _, name := range rs.Tags {
		targets = append(targets, state.TargetItem{
			Kind: state.TargetKindTag,
			Name: name,
			Ref:  name,
		})
	}
	if len(targets) == 0 {
		for _, name := range rs.Branches {
			targets = append(targets, state.TargetItem{
				Kind: state.TargetKindLocal,
				Name: name,
				Ref:  name,
			})
		}
	}
	if len(targets) == 0 {
		return state.New().WithBlocked(
			state.BlockTargetEmpty,
			"No branch targets available.",
			"Create or fetch a branch before merging, rebasing, or resetting.",
		)
	}

	status := state.New().WithTargetPick(action, targets)
	status.Message = "Use up/down to choose a target."
	status.Detail = "Enter previews the result. Esc returns to browse."
	return status
}
```

```go
func buildActionPreview(action state.Action, target string, rs git.Status, currentOnly, targetOnly int) state.Status {
	head := shorten(rs.Head, 12)
	switch action {
	case state.ActionMerge:
		switch {
		case currentOnly == 0 && targetOnly == 0:
			return state.New().WithOutcome(
				state.ActionMerge,
				"Target already matches HEAD.",
				"Nothing moves. The branch already points at the same commit.",
				true,
			)
		case currentOnly == 0:
			return state.New().WithOutcome(
				state.ActionMerge,
				"FF 가능. 포인터만 이동합니다.",
				"HEAD can move to "+target+". Current-only: "+fmt.Sprint(currentOnly)+"  Target-only: "+fmt.Sprint(targetOnly),
				true,
			)
		case targetOnly == 0:
			return state.New().WithOutcome(
				state.ActionMerge,
				"대상은 이미 포함되어 있습니다.",
				"Current branch already contains "+target+". Current-only: "+fmt.Sprint(currentOnly)+"  Target-only: "+fmt.Sprint(targetOnly),
				true,
			)
		default:
			return state.New().WithOutcome(
				state.ActionMerge,
				"FF 불가. merge commit이 필요합니다.",
				"HEAD "+head+" and target "+target+" have diverged. Current-only: "+fmt.Sprint(currentOnly)+"  Target-only: "+fmt.Sprint(targetOnly),
				true,
			)
		}
	case state.ActionRebase:
		// rebase preview branches
	case state.ActionReset:
		// reset preview branch
	default:
		return state.New().WithOutcome(action, "No action selected.", target, false)
	}
}
```

### AFTER
`actionPickTargets`는 흐름 조립만 담당하고, target 수집 규칙은 `navigation.go`와 공유하는 helper가 담당한다.

```go
func actionPickTargets(rs git.Status, action state.Action) state.Status {
	if blocksDetachedTargetAction(rs, action) {
		return detachedTargetBlockedStatus()
	}

	targets := buildActionTargetItems(rs)
	if len(targets) == 0 {
		return emptyTargetBlockedStatus()
	}

	status := state.New().WithTargetPick(action, targets)
	status.Message = "Use up/down to choose a target."
	status.Detail = "Enter previews the result. Esc returns to browse."
	return status
}

func blocksDetachedTargetAction(rs git.Status, action state.Action) bool {
	return rs.Detached && (action == state.ActionMerge || action == state.ActionRebase)
}

func buildActionTargetItems(rs git.Status) []state.TargetItem {
	targets := make([]state.TargetItem, 0, len(rs.LocalBranches)+len(rs.RemoteBranches)+len(rs.Tags))
	targets = appendLocalTargets(targets, rs)
	targets = appendRemoteTargets(targets, rs.RemoteBranches)
	targets = appendTagTargets(targets, rs.Tags)
	if len(targets) == 0 {
		targets = appendFallbackBranchTargets(targets, rs.Branches)
	}
	return targets
}

func appendLocalTargets(targets []state.TargetItem, rs git.Status) []state.TargetItem {
	for _, name := range rs.LocalBranches {
		upstream, known := branchUpstream(rs, name)
		targets = append(targets, state.TargetItem{
			Kind:       state.TargetKindLocal,
			Name:       name,
			Ref:        name,
			NoUpstream: known && upstream == "",
		})
	}
	return targets
}

func appendRemoteTargets(targets []state.TargetItem, remoteBranches []string) []state.TargetItem {
	for _, name := range remoteBranches {
		if isRemoteHeadRef(name) {
			continue
		}
		targets = append(targets, state.TargetItem{
			Kind: state.TargetKindRemote,
			Name: name,
			Ref:  name,
		})
	}
	return targets
}

func appendTagTargets(targets []state.TargetItem, tags []string) []state.TargetItem {
	for _, name := range tags {
		targets = append(targets, state.TargetItem{
			Kind: state.TargetKindTag,
			Name: name,
			Ref:  name,
		})
	}
	return targets
}

func appendFallbackBranchTargets(targets []state.TargetItem, branches []string) []state.TargetItem {
	for _, name := range branches {
		targets = append(targets, state.TargetItem{
			Kind: state.TargetKindLocal,
			Name: name,
			Ref:  name,
		})
	}
	return targets
}

func isRemoteHeadRef(name string) bool {
	return strings.HasSuffix(name, "/HEAD")
}
```

`navigation.go`의 `sectionTargets`도 같은 helper를 재사용하도록 옮긴다. 이렇게 해야 branch, remote, tag 규칙이 한 곳에서 유지된다.

```go
func sectionTargets(rs git.Status, section graphSection) []state.TargetItem {
	switch section {
	case sectionCurrent:
		return buildCurrentSectionTargets(rs)
	case sectionRemote:
		return buildRemoteSectionTargets(rs)
	case sectionTags:
		return buildTagSectionTargets(rs)
	default:
		return nil
	}
}
```

`buildActionPreview`는 action dispatch만 담당하고, 각 preview 생성은 별도 함수로 분리한다. preview가 더 자라면 파일도 나눈다.

```go
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
		return state.New().WithOutcome(
			state.ActionMerge,
			"Target already matches HEAD.",
			"Nothing moves. The branch already points at the same commit.",
			true,
		)
	case currentOnly == 0:
		return state.New().WithOutcome(
			state.ActionMerge,
			"FF 가능. 포인터만 이동합니다.",
			"HEAD can move to "+target+". "+countDetail(currentOnly, targetOnly),
			true,
		)
	case targetOnly == 0:
		return state.New().WithOutcome(
			state.ActionMerge,
			"대상은 이미 포함되어 있습니다.",
			"Current branch already contains "+target+". "+countDetail(currentOnly, targetOnly),
			true,
		)
	default:
		return state.New().WithOutcome(
			state.ActionMerge,
			"FF 불가. merge commit이 필요합니다.",
			"HEAD "+shorten(rs.Head, 12)+" and target "+target+" have diverged. "+countDetail(currentOnly, targetOnly),
			true,
		)
	}
}

func buildRebasePreview(target string, currentOnly, targetOnly int) state.Status {
	switch {
	case currentOnly == 0 && targetOnly == 0:
		return state.New().WithOutcome(
			state.ActionRebase,
			"Target already matches HEAD.",
			"Nothing is rewritten because both refs point at the same commit.",
			true,
		)
	case targetOnly == 0:
		return state.New().WithOutcome(
			state.ActionRebase,
			"Target is already in the base history.",
			"Current commits will replay onto "+target+". "+countDetail(currentOnly, targetOnly),
			true,
		)
	default:
		return state.New().WithOutcome(
			state.ActionRebase,
			"새 base 위로 커밋을 재배치합니다.",
			countDetail(currentOnly, targetOnly)+"  |  target: "+target,
			true,
		)
	}
}

func buildResetPreview(target string, rs git.Status, currentOnly, targetOnly int) state.Status {
	return state.New().WithOutcome(
		state.ActionReset,
		"현재 HEAD를 선택한 위치로 이동합니다.",
		"HEAD "+shorten(rs.Head, 12)+" -> "+target+"  |  "+countDetail(currentOnly, targetOnly),
		true,
	)
}

func countDetail(currentOnly, targetOnly int) string {
	return "Current-only: " + fmt.Sprint(currentOnly) + "  Target-only: " + fmt.Sprint(targetOnly)
}
```

`executionDetail`과 `findRemoteCommitHash`는 `actions.go`에 남기지 않고 별도 책임으로 분리한다.

```go
func executionDetail(action state.Action, target string, rs git.Status) string {
	switch action {
	case state.ActionPull:
		return "Upstream pointer is now reflected in the local branch."
	case state.ActionMerge:
		return "Merge complete. HEAD now reflects " + emptyDash(rs.Branch) + " with target " + target + "."
	case state.ActionRebase:
		return "Rebase complete. The branch was replayed on top of " + target + "."
	case state.ActionReset:
		return "Hard reset complete. HEAD now points at " + target + "."
	default:
		return "Action complete."
	}
}

func findRemoteCommitHash(rs git.Status, upstream string) string {
	if upstream == "" {
		return ""
	}
	target := normalizeRemoteUpstream(upstream)
	for _, commit := range rs.GraphCommits {
		if commitHasRemoteDecoration(commit, target) {
			return commit.Hash
		}
	}
	return ""
}
```

### Tests
리팩토링 전에 아래 테스트를 먼저 추가한다.

```go
func TestDeriveStatus(t *testing.T) { /* ... */ }
func TestActionPickTargetsBuildsSelectableTargets(t *testing.T) { /* ... */ }
func TestBuildActionPreview(t *testing.T) { /* ... */ }
func TestFindRemoteCommitHash(t *testing.T) { /* ... */ }
```

### Verification
```sh
go test ./internal/app
go test ./...
go build ./cmd/graphkeeper
```

`scripts/check`가 repo에 생기면 그 명령을 우선 사용한다.

### Notes
- 첫 단계에서는 message text를 바꾸지 않는다.
- `pullReady`가 `RebaseInProgress`를 보지 않는 현재 동작은 테스트로 드러낸 뒤 별도 수정으로 다룬다.
- 순수 함수 테스트를 우선하고, 실제 Git repository를 만드는 integration test는 필요한 경우에만 추가한다.
- target assembly는 `actions.go`와 `navigation.go`가 같이 쓰는 공용 policy로 본다.
- preview가 더 복잡해지면 `*_preview.go`로 파일을 나눈다.
- execution detail, remote commit lookup는 별도 파일로 분리해도 된다.
