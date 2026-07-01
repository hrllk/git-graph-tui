# Branch Delete Local / Origin Plan

## 목적

`graphkeeper` 에서 브랜치 삭제 기능을 추가한다.

이번 계획의 핵심은 다음 두 가지다.

1. local branch 삭제를 추가한다.
2. `origin` remote branch 삭제를 추가한다.

이 기능은 destructive action 이므로, 기존 실행형 command 흐름과 같은 `confirm -> execute -> loading toast -> result` 구조를 따른다.

## 참조 문서

이 문서는 다음 최근 문서를 기준으로 작성한다.

- `docs/archive/20260630-0004-implement-confirm-command-ux-plan.md`
- `docs/archive/20260630-0001-implement-progress-toast-plan-d.md`
- `docs/archive/20260701-0001-graph-actions-availability-plan.md`
- `docs/archive/20260701-0002-local-pane-split-layout-plan.md`
- `docs/structure.md`
- `docs/model-refactor-plan.md`

## 현재 관찰된 구현 상태

현재 코드에는 브랜치 생성과 checkout, merge, rebase, reset, push 계열 흐름은 있지만 브랜치 삭제는 없다.

- `state.Action` 에 delete 전용 action 이 없다.
- `renderActionHelpLines()` 에 delete shortcut 이 없다.
- `handleBrowseGraphKey()` 와 `handleBrowseSectionKey()` 에 branch delete 진입점이 없다.
- `internal/git/` 에는 local branch 삭제와 `origin` remote branch 삭제를 분리해 감싸는 helper 가 없다.

즉, 이번 작업은 단순히 git 명령 하나를 추가하는 문제가 아니라, `Graph` 액션 노출, confirm 문구, 실행 중 상태, 실패 처리까지 같이 묶어야 한다.

## 핵심 판단

### 1. local 삭제와 origin 삭제는 같은 UX 계열이지만 다른 command 다

삭제 행위는 공통 confirm UX 를 쓰되, 실제 실행 명령은 분리한다.

- local branch 삭제: `git branch -d <branch>`
- origin branch 삭제: `git push origin --delete <branch>`

둘을 하나의 버튼으로 합치지 않고, 브랜치 종류와 선택된 target 에 따라 분기하는 편이 안전하다.

### 2. delete 는 기본적으로 current branch 에서 차단해야 한다

삭제는 destructive action 이고, 현재 체크아웃된 local branch 는 삭제 대상이 될 수 없다.

권장 원칙:

- 현재 checkout 중인 local branch 는 삭제를 막는다.
- detached HEAD 상태에서는 current branch 삭제 규칙을 따로 처리한다.
- remote branch 삭제는 local checkout 상태와 무관하게 처리할 수 있지만, target ref 가 실제 `origin` 아래인지 먼저 확인한다.

### 3. 이 기능은 `origin` 에 한정한다

이번 계획의 범위는 remote 전체 삭제가 아니라 `origin` remote branch 삭제다.

다른 remote 이름까지 일반화하면 confirm 문구와 target 판정이 복잡해지므로, 첫 구현은 `origin` 으로 제한한다.

필요하면 후속 문서에서 `upstream remote` 일반화나 다중 remote 삭제를 분리한다.

### 4. force delete 는 이번 범위에 넣지 않는다

local branch 가 unmerged 인 경우의 force delete 는 별도 안전 정책이 필요하다.

이번 계획에서는 다음 원칙을 우선 둔다.

- `git branch -d` 의 안전 삭제를 기본으로 한다.
- unmerged local branch 는 실패 메시지로 되돌린다.
- 강제 삭제 `-D` 는 후속 문서로 분리한다.

## 범위

### 포함

- local branch 삭제
- `origin` remote branch 삭제
- delete 대상이 local / remote 인지에 따른 분기
- confirm 문구 추가
- execution toast 및 result state 정리
- Graph / Local 패널의 delete action 노출
- current branch 및 invalid target 차단
- 관련 테스트 추가

### 제외

- force delete `git branch -D`
- `origin` 이 아닌 다른 remote delete 일반화
- branch rename
- branch merge 정책 변경
- UI 레이아웃 재설계

## 상태 모델 방향

현재 `state.Status` 와 `Action` 구조를 그대로 확장하는 방향이 좋다.

권장 추가 항목:

```go
type Action string

const (
    ActionDeleteBranch Action = "delete-branch"
)
```

삭제는 target 종류에 따라 같은 action 을 쓰되, 실제 실행 단계에서 local / remote 를 구분한다.

예시 방향:

- `local branch` target 이면 `git branch -d`
- `origin/<name>` target 이면 `git push origin --delete`

확인 화면은 공통 `ModeConfirm` 을 재사용하고, 실행 중에는 공통 loading toast 를 사용한다.

## 구현 방향

### 1. target 판정을 먼저 고정한다

삭제 대상은 사용자가 현재 가리키는 항목에 따라 달라진다.

권장 규칙:

- `Local` 섹션의 branch 는 local delete 후보로 본다.
- `Remote` 섹션의 `origin/...` branch 는 remote delete 후보로 본다.
- tag, commit, HEAD, current checkout branch 는 삭제 대상이 아니다.

### 2. local delete 와 remote delete 를 분리한 helper 를 둔다

명령 실행은 한 함수에 몰지 말고, 대상별 helper 로 나눈다.

예시 방향:

```go
func deleteLocalBranch(repo *git.Repo, branch string) tea.Cmd
func deleteOriginBranch(repo *git.Repo, branch string) tea.Cmd
```

또는 내부적으로는 공통 실행 helper 를 두고, command 만 분리해도 된다.

중요한 점은 다음이다.

- target 판정은 UI 계층에서 한다.
- 실제 git command 는 execute 계층에서 한다.
- 실패 시 상태 복귀는 기존 `update_execute.go` 정책을 따른다.

### 3. confirm 문구는 짧고 명확하게 한다

삭제는 되돌리기 어렵기 때문에, confirm 에서 대상과 범위를 분명히 보여줘야 한다.

권장 문구:

- `Delete local branch?`
- `Delete origin branch?`
- `Delete branch <name>?`

detail 에는 다음을 짧게 넣는다.

- local 인지 origin 인지
- 현재 브랜치인지 여부
- 삭제 후 복구가 어렵다는 점

예시:

- `This will delete local branch feature/foo.`
- `This will remove origin/feature/foo from the remote.`
- `Current branch cannot be deleted.`

### 4. 실행 중에는 공통 toast 를 사용한다

삭제 command 도 다른 execution flow 와 동일하게 실행 중 상태를 보여준다.

예시:

- `Deleting branch...`
- `Deleting local branch...`
- `Deleting origin branch...`

toast 는 결과 메시지와 역할이 겹치지 않도록 짧게 유지한다.

### 5. result state 는 실패 이유를 숨기지 않는다

local delete 실패는 보통 다음 이유 중 하나다.

- 현재 브랜치임
- unmerged branch 임
- 브랜치가 존재하지 않음

remote delete 실패는 보통 다음 이유 중 하나다.

- `origin` branch 가 없음
- remote 가 없음
- 네트워크 또는 auth 실패

이런 경우에는 단순히 빈 화면으로 넘기지 말고, `Blocked` 또는 `Error` 상태로 이유를 보여준다.

## UI 노출 방향

### Graph 액션

`Graph` 섹션에는 delete action 을 추가한다.

권장 형태:

- `• d: delete branch`
- local branch 에서만 활성화
- remote branch 에서도 별도 target 으로 활성화

`Graph Actions Availability Plan` 의 원칙에 맞춰, 목록 자체는 고정하고 활성 / 비활성만 바꾸는 편이 좋다.

### Local / Remote 패널

우측 패널의 branch 목록에서 삭제 가능 여부를 함께 보여준다.

권장 원칙:

- local branch 는 현재 checkout branch 일 때 비활성 처리한다.
- remote branch 는 `origin` 에 속하는 branch 만 활성화한다.
- 설명은 길게 늘리지 말고 `current branch`, `origin only`, `not deletable` 정도로 짧게 유지한다.

## 파일 경계

기존 구조를 크게 흔들지 않는 방향이 좋다.

권장 수정 후보:

- `internal/state/state.go`
- `internal/app/key_handling_browse.go`
- `internal/app/key_handling_confirm.go`
- `internal/app/update_execute.go`
- `internal/app/view_sections.go`
- `internal/app/view_shell.go`
- `internal/app/commands.go`
- `internal/app/actions.go`
- `internal/git/repo.go`
- `internal/git/repo_exec.go`
- `internal/app/commands_test.go`
- `internal/app/key_handling_test.go`
- `internal/app/model_test.go`

## 구현 순서

1. 삭제 대상 규칙을 먼저 정리한다.
2. `ActionDeleteBranch` 를 추가한다.
3. local / origin 삭제 helper 를 만든다.
4. browse / confirm 키 흐름에 delete 진입점을 추가한다.
5. confirm 문구와 loading toast 를 정리한다.
6. 실패 시 blocked / error 상태를 연결한다.
7. Graph / Local / Remote UI 에서 delete action 을 노출한다.
8. 관련 테스트를 추가한다.

## 테스트 전략

삭제 기능은 회귀 위험이 높으므로, command 실행과 UI 가 같이 맞물리는지 확인해야 한다.

우선 확인할 항목은 다음과 같다.

- current branch 는 local delete 대상에서 제외되는지
- local branch delete 가 `git branch -d` 로 실행되는지
- unmerged local branch 삭제 실패가 명확히 표시되는지
- `origin/<name>` 삭제가 `git push origin --delete` 로 실행되는지
- `origin` 이 아닌 remote 대상이 차단되는지
- confirm 승인 후 loading toast 가 표시되는지
- 취소 시 browse / previous state 로 복귀하는지
- delete action 이 Graph / Local / Remote UI 에서 일관되게 보이는지

권장 테스트 파일:

- `internal/app/key_handling_test.go`
- `internal/app/commands_test.go`
- `internal/app/update_execute_test.go`
- `internal/app/model_test.go`
- `internal/app/view_sections_test.go`
- `internal/git/repo_test.go`

## 검증

```sh
go test ./internal/app
go test ./internal/git
go test ./...
go build ./cmd/graphkeeper
```

## 완료 기준

- local branch 삭제가 가능하다.
- `origin` remote branch 삭제가 가능하다.
- current branch 삭제는 차단된다.
- 삭제 실패 사유가 사용자에게 명확히 보인다.
- confirm 과 loading toast 가 다른 실행형 command 와 같은 규칙으로 동작한다.
- 기존 branch create / checkout / merge / rebase / reset 흐름이 회귀하지 않는다.

## 메모

이 문서는 branch delete 기능의 UX 와 구현 경계를 먼저 고정하는 계획이다.

force delete, 다른 remote 일반화, bulk delete 는 별도 문서로 분리하는 편이 안전하다.
