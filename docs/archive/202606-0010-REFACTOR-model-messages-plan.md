# `internal/app/model.go` 메시지 타입 분리 계획서

## 목표

Bubble Tea message 타입을 `model.go`에서 떼어내어, `commands.go` / `update.go`가 쓰는 전송 전용 타입을 별도 파일로 모은다.

핵심은 상태 정의 파일과 비동기 메시지 타입을 분리하는 것이다.

## 범위

- 대상 파일
  - `internal/app/model.go`
- 후보 파일
  - `internal/app/messages.go`
  - `internal/app/update_messages.go`
- 관련 파일
  - `internal/app/commands.go`
  - `internal/app/update.go`
- 관련 테스트
  - `internal/app/commands_test.go`
  - `internal/app/model_test.go`
- 새 package 생성: 하지 않음

## 참조 문서

이 계획은 다음 문서의 분리 원칙을 이어받는다.

- `docs/archive/202606-0002-REFACTOR-commands-test-plan.md`
- `docs/archive/202606-0007-REFACTOR-key-handling-plan.md`
- `docs/model-refactor-plan.md`

## 현재 문제

`model.go`에는 다음 메시지 타입이 섞여 있다.

- `loadedMsg`
- `refreshedMsg`
- `fetchedMsg`
- `preparedMsg`
- `pullCheckedMsg`
- `previewMsg`
- `executedMsg`
- `createdBranchMsg`
- `pullFetchedMsg`
- `pushFetchedMsg`
- `pullPreviewReadyMsg`

이 타입들은 상태 정의가 아니라, command/update 경계에서 쓰는 전달 객체다.

## 분리 원칙

### `model.go`에 남길 것

- `model` 구조체
- `graphSection`
- `New()`
- `Init()`

### 메시지 파일로 옮길 것

- command 결과 메시지
- async 실행 결과 메시지
- pull preview / branch creation 결과 메시지

## 구현 포인트

1. 메시지 타입을 `messages.go` 또는 `update_messages.go`로 옮긴다.
2. `commands.go` / `update.go`의 import를 정리한다.
3. `commands_test.go`와 `model_test.go`에서 직접 참조하는 메시지 타입 import와 타입 단언을 새 파일 기준으로 유지한다.
4. 메시지 이름은 현재 의미를 유지하되, 상태 파일과는 분리한다.

## 검증 기준

1. `model.go`에 message type 선언이 남아 있지 않다.
2. `commands.go`와 `update.go`가 새 메시지 파일을 자연스럽게 참조한다.
3. command/update 동작은 바뀌지 않는다.

## 비고

- 메시지 타입은 기능이 아니라 전달 계약이다.
- 상태 정의 파일에 두면 `model.go`가 다시 비대해진다.
