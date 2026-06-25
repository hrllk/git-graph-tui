# `internal/app` 그래프 리팩토링 아카이브 인덱스

## 용도

이 문서는 `view`, `navigation`, `graph_render` 관련 리팩토링 계획을 묶어 둔 아카이브 인덱스다.

현재 기준의 참조 문서는 아래를 우선으로 본다.

- `docs/structure.md`
- `docs/roadmap.md`
- `docs/model-refactor-plan.md`

## 관련 아카이브 문서

- [202606-0004-REFACTOR-view-graph-structure-plan.md](./202606-0004-REFACTOR-view-graph-structure-plan.md)
- [202606-0005-REFACTOR-navigation-graph-rules-plan.md](./202606-0005-REFACTOR-navigation-graph-rules-plan.md)
- [202606-0006-REFACTOR-graph-render-core-plan.md](./202606-0006-REFACTOR-graph-render-core-plan.md)

## 현재 해석

이 묶음은 과거에 `internal/app` 안의 그래프 렌더링과 navigation 경계를 정리하려고 작성된 계획이다.

이후 구조가 바뀌면서 일부 전제는 더 이상 유효하지 않다.

- DAG 기반 lazy load 전제는 폐기되었다.
- 그래프는 `git log` 기반 row 렌더링을 기준으로 본다.
- `model.go` 축소는 별도 문서로 분리했다.

## 읽는 순서

1. `docs/structure.md`로 현재 파일 배치를 확인한다.
2. `docs/roadmap.md`로 다음 작업 순서를 확인한다.
3. `docs/model-refactor-plan.md`로 `model.go` 축소 기준을 확인한다.
4. 아래 아카이브 문서들은 과거 설계 의도와 경계 정리를 참고할 때만 본다.

## 비고

- 이 문서 묶음은 구현 기준서라기보다 히스토리 기록에 가깝다.
- 현재 구현 기준은 아카이브가 아니라 `docs/structure.md`와 `docs/roadmap.md`다.
