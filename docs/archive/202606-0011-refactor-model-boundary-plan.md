# `internal/app/model.go` 헬퍼/상태 경계 정리 계획서

## 목표

`model.go`에 남아 있는 범용 헬퍼와 상태 경계 설명을 정리하고, `internal/state/state.go`와의 차이를 구현 단계에서 함께 고정한다.

핵심은 `model`과 `state.Status`가 같은 층이 아니라는 점을 분명히 하는 것이다.

## 범위

- 대상 파일
  - `internal/app/model.go`
- 관련 파일
  - `internal/state/state.go`
  - `internal/app/view.go`
  - `internal/app/execution_detail.go`
  - `internal/app/preview.go`
- 관련 테스트
  - `internal/app/model_test.go`
  - `internal/app/preview_test.go`
- 새 package 생성: 하지 않음

## 참조 문서

이 계획은 다음 문서의 경계 설명을 이어받는다.

- `docs/archive/architecture.md`
- `docs/archive/cli-structure-plan.md`
- `docs/model-refactor-plan.md`

## 현재 문제

`model.go`에는 범용 헬퍼와 상태 관련 상수가 섞여 있다.

- `min`
- `emptyDash`
- `shorten`
- `graphNode`, `laneSide`, `laneRef`, `graphRow` alias 일부

이 중 일부는 app 내부 렌더링/판정에서 재사용되지만, 상태 정의 파일에 그대로 둘 필요는 없다.

## `model.go`와 `internal/state/state.go`의 차이

### `internal/state/state.go`

`state.go`는 UI가 보여줄 상태를 정의한다.

즉, 다음을 담는다.

- 화면 모드
- 액션 종류
- 차단 사유
- 타겟 선택 정보
- 사용자에게 보여줄 제목/메시지/상세

`state.Status`는 도메인 수준의 UI 상태 모델이다.

### `internal/app/model.go`

`model.go`는 Bubble Tea 앱의 실행 컨테이너다.

즉, 다음을 담는다.

- repo 핸들
- 현재 repo 상태 스냅샷
- cursor / scroll / layout 정보
- branch input 상태
- async 처리용 내부 플래그

`model`은 `state.Status`를 포함하지만, `state.Status`와 같은 층은 아니다.

## 분리 원칙

### `state.go`에 남는 성격

- 사용자에게 보여줄 상태
- mode / action / block / target selection

### `model.go`에 남는 성격

- 앱 실행 상태
- repository snapshot
- cursor and scroll
- temporary input state

### 다른 파일로 빼도 되는 헬퍼

- 렌더링용 문자열 trimming
- 표시용 기본값 치환
- 최소값 계산처럼 범용성이 높은 함수

## 구현 포인트

1. `model.go`에 있는 범용 헬퍼를 의미에 맞는 파일로 이동한다.
2. alias 중 앱 실행 상태와 직접 관련 없는 것은 축소한다.
3. `state.Status`와 `model`의 책임이 겹치지 않도록 설명을 남긴다.

## 검증 기준

1. `state.Status`는 UI 상태 모델로 읽힌다.
2. `model`은 Bubble Tea 실행 컨테이너로 읽힌다.
3. 범용 헬퍼가 상태 정의 파일의 핵심 책임처럼 보이지 않는다.
4. 렌더링과 preview 결과는 바뀌지 않는다.

## 비고

- 이 문서는 "왜 분리해야 하는가"를 설명하는 문서이면서, 실제 구현 시 경계 판단 기준으로 함께 반영한다.
- 구현은 0009와 0010에 이어 진행하되, 이 문서의 경계 정의도 코드와 함께 맞춘다.
