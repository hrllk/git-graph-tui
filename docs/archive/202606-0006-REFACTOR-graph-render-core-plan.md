# `internal/app/graph_render.go` core 정리 계획서

## 상태

이 문서는 graph row, connector, compact format을 분리하려던 과거 계획이다.

현재 기준은 `docs/structure.md`, `docs/roadmap.md`, `docs/model-refactor-plan.md`다.

## 당시 의도

- row 렌더링과 포맷 로직을 분리한다.
- connector 렌더링을 별도 파일로 뺀다.
- `graph_render.go`는 orchestration 중심으로 줄인다.

## 현재 해석

이 계획의 방향성은 여전히 유효하다.

다만 실제 분리 순서는 navigation 규칙과 view 조립 경계가 먼저 안정된 뒤에 다시 잡는 편이 안전하다.

## 관련 문서

- `docs/structure.md`
- `docs/roadmap.md`
- `docs/model-refactor-plan.md`
- `docs/archive/202606-0005-REFACTOR-navigation-graph-rules-plan.md`

## 비고

- 이 문서는 구현 완료 기준서가 아니라 과거의 분리 순서를 기록한 문서다.
- 현재 렌더링 구조가 바뀌었다면 helper 이름보다 책임 경계를 먼저 다시 봐야 한다.
