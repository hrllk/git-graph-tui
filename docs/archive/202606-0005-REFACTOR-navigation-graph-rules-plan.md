# `internal/app/navigation.go` 중복 규칙 분리 계획서

## 상태

이 문서는 navigation과 graph decoration 규칙의 중복을 줄이기 위해 작성된 계획이다.

현재 참조 기준은 `docs/structure.md`, `docs/roadmap.md`, `docs/model-refactor-plan.md`다.

## 당시 핵심 문제

- cursor/page 이동과 branch 판정 규칙이 한 파일에 섞여 있었다.
- 렌더링과 navigation이 같은 decoration 규칙을 따로 해석할 위험이 있었다.

## 현재 해석

핵심 방향은 여전히 유효하다.

- `navigation.go`는 cursor와 browse state 중심으로 유지한다.
- branch/decorations 판정 규칙은 가능한 한 공용 helper로 모은다.
- `isLocalGraphPointer` 같은 판정은 렌더링과 같은 기준을 봐야 한다.

## 관련 문서

- `docs/structure.md`
- `docs/roadmap.md`
- `docs/model-refactor-plan.md`
- `docs/archive/202606-0006-REFACTOR-graph-render-core-plan.md`

## 비고

- 이 문서는 구현 체크리스트라기보다 규칙 분리의 배경 설명에 가깝다.
- 실제 helper 분리 여부는 현재 `internal/app`의 파일 배치와 테스트로 다시 확인해야 한다.
