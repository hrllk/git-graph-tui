# `internal/app/view.go` / `internal/app/view_graph.go` 분리 계획서

## 상태

이 문서는 `view.go`에서 그래프 pane 조립을 분리하려던 과거 계획이다.

현재 참조 기준은 `docs/structure.md`, `docs/roadmap.md`, `docs/model-refactor-plan.md`다.

## 당시 의도

- `view.go`는 전체 프레임만 조립한다.
- 그래프 pane 조립은 별도 파일로 뺀다.
- detail, footer, popup과 그래프 조립의 경계를 분리한다.

## 현재 해석

이 방향성은 여전히 유효하지만, 실제 작업 여부는 현재 파일 배치와 테스트 상태를 기준으로 다시 판단해야 한다.

특히 `renderDetailContent`, `renderStatusCompact`, `renderTargets`는 그래프 pane와 무조건 같이 움직이지 않는다.

## 관련 문서

- `docs/structure.md`
- `docs/roadmap.md`
- `docs/model-refactor-plan.md`
- `docs/archive/202606-0006-REFACTOR-graph-render-core-plan.md`

## 비고

- 이 문서는 완료된 설계가 아니라 과거의 분리 의도를 남긴 아카이브다.
- 현재 구현은 이 계획보다 더 단순하거나 더 분리되어 있을 수 있으므로, 실제 코드와 함께 읽어야 한다.
