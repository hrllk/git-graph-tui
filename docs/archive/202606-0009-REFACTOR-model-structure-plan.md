# `internal/app/model.go` 상태 정의 정리 계획서

## 목표

`internal/app/model.go`에서 런타임 상태 정의만 남기고, lazy-load 흔적과 실사용이 약한 상수를 정리한다.

핵심은 `model.go`를 Bubble Tea 앱의 컨테이너로만 유지하는 것이다.

## 범위

- 대상 파일
  - `internal/app/model.go`
- 관련 파일
  - `internal/app/navigation.go`
  - `internal/app/update.go`
  - `internal/app/commands.go`
  - `internal/app/view.go`
- 관련 테스트
  - `internal/app/model_test.go`
  - `internal/app/navigation_test.go`가 있으면 함께 확인
- 새 package 생성: 하지 않음

## 참조 문서

이 계획은 다음 문서의 구조를 이어받는다.

- `docs/archive/cli-structure-plan.md`
- `docs/archive/architecture.md`
- `docs/archive/202606-0003-REFACTOR-graph-render-test-plan.md`
- `docs/model-refactor-plan.md`

## 현재 문제

현재 `model.go`에는 상태 정의 외의 요소가 함께 들어 있다.

- `initialGraphCommitLimit`
- `graphLoadIncrement`
- `graphLoadThreshold`
- `sectionLocal`
- `graphSection`과 상태 정의가 아닌 alias 일부

이 요소들은 현재 구조에서 책임이 불명확하다.

특히 graph lazy load 상수는 더 이상 사용하지 않으므로 삭제 대상이다.
`sectionLocal`도 동일하게 더 이상 쓰지 않는 섹션 잔재이므로 제거한다.

## 분리 원칙

### `model.go`에 남길 것

- `model` 구조체
- `graphSection` 정의
- `New(repo *git.Repo)`
- `Init() tea.Cmd`
- 앱 실행 상태를 유지하는 데 필요한 최소 필드

### 정리 대상

- lazy-load 상수
- 사용되지 않는 section/state 잔재
- 모델 상태와 직접 관계 없는 상수

## 구현 포인트

1. `model` 필드 중 실제로 앱 실행 상태를 나타내는 것과 그렇지 않은 것을 나눈다.
2. lazy-load 상수를 삭제한다.
3. `sectionLocal`을 제거한다.
4. `model.go`가 상태 정의 파일로 읽히도록 구조를 압축한다.

## 검증 기준

1. `model.go`를 열었을 때 앱 상태 정의가 바로 보인다.
2. lazy-load 상수가 남아 있지 않다.
3. 상태 정의와 UI 상태 모델이 겹치지 않는다.
4. graph 렌더링 결과는 바뀌지 않는다.

## 비고

- 이 문서는 메시지 타입 분리와는 분리해서 본다.
- 상태 정의가 작아져야 뒤 단계에서 메시지와 헬퍼 분리가 쉬워진다.
