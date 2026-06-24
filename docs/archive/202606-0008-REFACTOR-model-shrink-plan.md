# `internal/app/model.go` 리팩토링 마스터 계획서

## 목적

이 문서는 `internal/app/model.go` 리팩토링의 마스터 인덱스다.

`model.go` 축소 작업은 아래 3개 상세 문서로 나눈다.

1. `model.go`의 상태 정의와 lazy-load 정리
2. Bubble Tea message 타입 분리
3. 헬퍼/상태 경계 정리와 `internal/state/state.go` 차이 설명

## 하위 문서

- [202606-0009-REFACTOR-model-structure-plan.md](./202606-0009-REFACTOR-model-structure-plan.md)
- [202606-0010-REFACTOR-model-messages-plan.md](./202606-0010-REFACTOR-model-messages-plan.md)
- [202606-0011-REFACTOR-model-boundary-plan.md](./202606-0011-REFACTOR-model-boundary-plan.md)

## 적용 순서

1. `model.go` 상태 정의를 정리한다.
2. 메시지 타입을 별도 파일로 뺀다.
3. 헬퍼와 상태 경계를 정리한다.
4. 관련 테스트를 책임별로 재배치한다.
5. `docs/model-refactor-plan.md`를 최종 작업 노트로 유지한다.

## 완료 기준

- 각 상세 문서만 읽어도 해당 단계 구현이 가능하다.
- `model.go`의 책임이 명확하게 줄어든다.
- `internal/state/state.go`와의 경계가 문서로 분명해진다.

## 비고

- 이 문서는 구현 기준서라기보다 분해된 상세 문서의 목차다.
- 상세 구현 내용은 하위 문서를 먼저 본다.
