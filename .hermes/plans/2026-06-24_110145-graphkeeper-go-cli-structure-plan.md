# graphkeeper Go CLI 2026 구조화 계획서

> **For Hermes:** 이 계획은 task-by-task로 구현하기 위한 기준 문서다. 구현 시에는 각 단계마다 기존 동작을 유지하면서 구조만 먼저 분리한다.

**Goal:** `graphkeeper`를 단일 MVP 코드베이스에서 2026년 기준의 현대적인 Go CLI 구조로 재정렬해, 진입점, 도메인 로직, Git 어댑터, TUI 렌더링, 상태 관리, 테스트를 명확히 분리한다.

**Architecture:**
- `cmd/`는 얇은 bootstrap만 담당한다.
- `internal/`는 앱 전용 구현을 담고, `pkg/`는 당분간 만들지 않는다.
- Bubble Tea 기반 UI는 `internal/app` 또는 `internal/ui/tui`에 두되, 상태 머신과 렌더링을 분리한다.
- Git 명령 실행은 `internal/git` 같은 어댑터로 제한하고, commit graph/lane 계산은 별도 순수 로직 패키지로 분리한다.
- 구조화의 1차 목표는 기능 추가가 아니라 유지보수성과 테스트 가능성 확보다.
- lazygit은 구조 참고용 벤치마크로만 사용하고, 패키지 배치 자체는 graphkeeper 규모에 맞게 더 단순하게 가져간다.

**Tech Stack:** Go 1.25.x, Bubble Tea, Lip Gloss, standard library, git CLI, table-driven tests, package-level unit tests.

---

## 1. 현재 코드베이스 진단

현재 구조는 MVP로서는 충분히 작동하지만, 장기 유지보수 관점에서는 책임이 많이 섞여 있다.

### 현재 관찰된 구조

- `cmd/git-graph-tui/main.go`
  - Git repo 열기
  - app model 생성
  - Bubble Tea program 실행
- `internal/app/model.go`
  - 상태 머신
  - key handling
  - command scheduling
  - graph row 계산
  - branch/lane selection
  - action execution 메시지 타입
- `internal/app/view.go`
  - 전체 레이아웃
  - graph/detail/status rendering
- `internal/git/repo.go`
  - git command execution
  - repository status gathering
  - branch metadata / divergence / graph data
- `internal/state/state.go`
  - UI 상태, action, mode 관리
- `internal/telemetry/telemetry.go`
  - logging/diagnostics

### 구조적 문제

1. `model.go`가 너무 크다.
   - app state, update logic, graph logic, action orchestration이 한 파일에 몰려 있다.
2. Git 접근과 도메인 계산이 섞여 있다.
   - `repo.go`는 단순 adapter를 넘어 상태 해석까지 한다.
3. View와 domain 계산의 경계가 약하다.
   - 렌더링이 data shaping을 많이 담당한다.
4. entrypoint가 너무 얇긴 하지만, bootstrap 계층이 없다.
   - 나중에 설정/로그/키맵/feature flag를 넣기 어렵다.
5. 현재 binary/package naming이 제품명과 맞지 않는다.
   - repo는 `graphkeeper`인데 binary/package는 `git-graph-tui`다.
   - 사용자 경험과 release naming을 정리할 필요가 있다.

---

## 2. Go CLI에서 흔히 쓰는 구조 패턴 분석

현대적인 Go CLI는 보통 아래 3가지 패턴을 섞는다.

### 패턴 A: 단일 바이너리 + 얇은 main

가장 흔한 형태다.

- `cmd/<app>/main.go`
- `internal/...` 아래에 모든 구현
- main은 wiring만 수행

장점:
- 시작이 쉽고 구조가 단순하다.
- 테스트가 쉬워진다.
- MVP에서 과한 추상화를 피할 수 있다.

적합성:
- 현재 `graphkeeper`는 아직 이 패턴이 가장 맞다.

### 패턴 B: 서브커맨드 기반 CLI

예: `app serve`, `app status`, `app doctor`

보통 다음을 쓴다.

- Cobra
- urfave/cli
- stdlib flag + 수동 디스패치

장점:
- 명령 수가 많아질 때 좋다.
- 외부 자동화와 결합하기 쉽다.

적합성:
- 지금은 아직 과하다.
- TUI 중심 앱은 서브커맨드보다 인터랙션 구조가 더 중요하다.
- 단, 향후 `graphkeeper doctor`, `graphkeeper dump`, `graphkeeper config` 같은 보조 명령이 생기면 도입 가치가 있다.

### 패턴 C: adapter / service / domain 분리

현대 Go에서 가장 실용적인 구조다.

- adapter: git CLI, filesystem, network, UI
- domain: graph, branch, action rules
- service/usecase: 화면 상태 전환, action orchestration

장점:
- 테스트하기 쉽다.
- UI와 core 로직이 분리된다.
- 향후 Neovim integration이나 headless mode 추가가 쉬워진다.

적합성:
- `graphkeeper`의 장기 구조는 이 방향이 맞다.

### 2.1 lazygit 구조 벤치마크

lazygit은 같은 Go TUI 계열이지만, 현재 graphkeeper보다 훨씬 큰 제품이다.

- 핵심 코드가 `pkg/` 아래에 많이 모여 있다.
- `pkg/gui`가 UI 역할을 크게 담당한다.
- `pkg/commands`가 git/OS 실행을 담당한다.
- `pkg/config`, `pkg/theme`, `pkg/utils`처럼 책임별 분리가 이미 진행돼 있다.
- integration test와 스크립트/스키마가 제품 규모에 맞게 분리돼 있다.

graphkeeper에 대한 해석:

- lazygit의 “역할 분리 원칙”은 좋다.
- lazygit의 “pkg 중심 대규모 구조”는 MVP인 graphkeeper에는 과하다.
- 따라서 graphkeeper는 `internal/` 중심 구조를 유지하고, 필요한 경우에만 얇은 기능 패키지를 추가하는 편이 맞다.

---

## 3. 권장 목표 구조

아래는 1차 목표 구조다. 지금 당장 전부 만들기보다, 단계적으로 이 형태로 수렴한다.

주의: 이 구조는 lazygit처럼 `pkg/`에 넓게 펼치는 방식이 아니라, graphkeeper 규모에 맞춘 `internal/` 중심 구조다.

```text
cmd/
  graphkeeper/
    main.go

internal/
  bootstrap/
    app.go
  app/
    model.go
    update.go
    view.go
    commands.go
    navigation.go
    actions.go
  git/
    repo.go
    status.go
    graph.go
  graph/
    lane.go
    layout.go
    sort.go
  state/
    state.go
  ui/
    theme.go
    widgets.go
  telemetry/
    telemetry.go
  config/
    config.go   (필요할 때만)

docs/
  architecture.md
  roadmap.md
  cli-structure-plan.md
```

### 역할 분리 원칙

- `cmd/graphkeeper/main.go`
  - flag/args 파싱
  - config 로드
  - bootstrap 호출
  - program 시작
- `internal/bootstrap`
  - 의존성 조립
  - repo/app/tui wiring
- `internal/app`
  - Bubble Tea model
  - update loop
  - key binding과 command dispatch
- `internal/git`
  - git CLI 실행
  - repo status 조회
  - raw git 정보 수집
- `internal/graph`
  - commit graph, lane, ordering, selection rule
  - 가능한 한 순수 함수 위주
- `internal/ui`
  - 스타일/테마/공통 widget
  - rendering helper
- `internal/state`
  - 상태 enum / action enum / status model
- `internal/telemetry`
  - 로그와 진단

---

## 4. 이번 구조화에서 지켜야 할 원칙

### 4.1 `cmd/`는 얇게 유지

main은 아래만 하도록 제한한다.

- args/flags 읽기
- config 준비
- repo open
- app bootstrap
- 프로그램 실행

main에 business logic을 두지 않는다.

### 4.2 `pkg/`는 지금 만들지 않음

이 프로젝트는 아직 public library가 아니다.

- 외부 import를 허용할 이유가 없다.
- public API를 약속하기 전에 내부 구조를 안정화해야 한다.

### 4.3 UI와 domain 계산을 분리

예:

- graph lane sort
- commit focus selection
- branch target ranking
- action eligibility 판단

이런 로직은 렌더링 함수 안에 넣지 않는다.

### 4.4 git CLI 호출과 해석을 분리

- `git status`, `git log`, `git for-each-ref` 같은 실제 호출은 adapter
- 이를 조합해 `Status`를 만드는 것은 service 또는 graph layer

### 4.5 기능보다 안정성 우선

구조화 단계에서는 새 기능을 늘리지 않는다.

- checkout/branch/pull/reset UX를 바꾸기보다
- 기존 기능이 어디에 있어야 하는지 먼저 정리한다.

---

## 5. 추천 마이그레이션 계획

### Phase 0. 기준선 고정

**목표:** 구조를 바꾸기 전에 현재 동작을 안전하게 고정한다.

**작업:**
- 현재 `go test ./...` 결과를 기준선으로 기록한다.
- `internal/app/model_test.go`, `internal/git/repo_test.go`의 기존 커버리지를 확인한다.
- `zz_temp_*` 임시 테스트 파일은 구조화 작업에 포함하지 말고, 별도 정리 대상으로 둔다.

**검증:**
- `go test ./...`
- `go build ./cmd/git-graph-tui`

### Phase 1. 진입점 정리

**목표:** binary 이름과 bootstrap 경계를 정리한다.

**권장 변경:**
- `cmd/git-graph-tui/main.go`를 `cmd/graphkeeper/main.go`로 이동 또는 새 entrypoint 추가
- `internal/bootstrap/app.go` 생성
- main은 bootstrap 함수만 호출하도록 축소

**파일 후보:**
- Create: `cmd/graphkeeper/main.go`
- Create: `internal/bootstrap/app.go`
- Modify: `README.md`
- Optional: `go.mod` module path 검토

**검증:**
- `go build ./cmd/graphkeeper`
- 기존 binary 출력/동작 유지 확인

### Phase 2. Git adapter와 graph domain 분리

**목표:** `repo.go`의 책임을 줄이고, graph 계산을 순수 로직으로 이동한다.

**권장 분리:**
- `internal/git/repo.go`: command execution, raw status gathering
- `internal/git/status.go`: status DTO 정리
- `internal/graph/lane.go`: lane ordering/compaction
- `internal/graph/layout.go`: commit row generation
- `internal/graph/sort.go`: branch/commit priority rule

**파일 후보:**
- Create: `internal/graph/lane.go`
- Create: `internal/graph/layout.go`
- Create: `internal/graph/sort.go`
- Modify: `internal/git/repo.go`
- Modify: `internal/app/model.go`

**검증:**
- 새 graph 함수에 unit test 추가
- `go test ./internal/graph ./internal/git ./internal/app`

### Phase 3. Bubble Tea model 분해

**목표:** `model.go`를 역할별 파일로 나눠서 읽기 쉽게 만든다.

**권장 분할:**
- `model.go`: struct, constructors, shared helpers
- `update.go`: 메시지 처리와 key handling
- `commands.go`: background command 생성
- `navigation.go`: cursor/section 이동
- `actions.go`: execute/confirm/preview orchestration

**파일 후보:**
- Split `internal/app/model.go`
- Keep `internal/app/view.go`
- Add tests to `internal/app/model_test.go`

**검증:**
- `go test ./internal/app -v`
- `go test ./...`

### Phase 4. View / theme 정리

**목표:** 렌더링 유틸과 스타일을 분리한다.

**권장 분리:**
- `view.go`: 화면 조립
- `widgets.go`: 재사용 가능한 UI building blocks
- `theme.go`: lipgloss style set

**파일 후보:**
- Create: `internal/ui/theme.go`
- Create: `internal/ui/widgets.go`
- Modify: `internal/app/view.go`

**검증:**
- 화면 출력 스냅샷 또는 문자열 기반 테스트
- manual TUI check

### Phase 5. 문서와 운영경험 정리

**목표:** 사용자가 구조를 이해하고 확장할 수 있게 한다.

**권장 문서:**
- `docs/architecture.md`: 패키지 책임과 데이터 흐름
- `docs/roadmap.md`: 다음 단계 기능 우선순위
- `README.md`: quickstart와 directory overview 정리

**검증:**
- README의 binary name이 실제 빌드 결과와 일치하는지 확인
- 문서 링크가 깨지지 않는지 확인

---

## 6. 현재 프로젝트에 대한 구체적 제안

### 6.1 binary 이름을 정리

현재는 `git-graph-tui`가 실제 binary/package 이름처럼 보인다.

권장:
- 프로젝트 제품명은 `graphkeeper`
- 실행 파일도 `graphkeeper`
- 경로는 `cmd/graphkeeper`

이렇게 맞추면 release, docs, packaging이 단순해진다.

### 6.2 `internal/app/model.go`는 반드시 분리

이 파일은 앞으로 더 커질 가능성이 높다.

우선순위 분리 대상:
- graph 관련 helper
- command scheduling helper
- selection/cursor helper
- action preview helper
- message type 정의

### 6.3 graph 계산은 테스트 가능한 순수 함수로 이동

특히 아래는 `internal/graph`로 빼는 것이 좋다.

- lane priority
- collapse rule
- focus branch selection
- deterministic ordering
- display row generation

### 6.4 README는 기능 설명보다 구조 설명을 추가

README에 다음이 들어가면 좋다.

- binary name
- package map
- execution flow
- what belongs in each package
- current MVP scope

---

## 7. 테스트 전략

구조화는 코드 이동이 많아서 테스트가 핵심이다.

### 권장 테스트 계층

1. 순수 로직 테스트
   - `internal/graph/*_test.go`
   - lane/order/selection
2. adapter 테스트
   - `internal/git/*_test.go`
   - git fixture 기반
3. model/update 테스트
   - `internal/app/*_test.go`
   - key handling / command dispatch
4. integration smoke test
   - `go build ./cmd/graphkeeper`
   - 가능한 경우 repo fixture에서 TUI 없는 상태 검증

### 최소 검증 명령

```bash
go test ./...
go build ./cmd/graphkeeper
go test ./internal/app ./internal/git ./internal/graph -v
```

---

## 8. 리스크와 트레이드오프

### 리스크 1: 구조만 바꾸고 복잡도는 그대로 남는 경우

대응:
- 파일 분리만 하지 말고 책임도 같이 이동한다.
- 특히 graph 계산과 view rendering을 진짜로 분리한다.

### 리스크 2: 패키지를 너무 많이 쪼개는 경우

대응:
- 처음부터 10개 이상 패키지로 가지 않는다.
- `internal/app`, `internal/git`, `internal/graph`, `internal/ui`, `internal/state` 정도로 시작한다.

### 리스크 3: Cobra 같은 프레임워크를 성급히 도입하는 경우

대응:
- 현재는 단일 TUI binary이므로 stdlib + Bubble Tea 조합이 더 적합하다.
- 서브커맨드가 실제로 필요해질 때만 프레임워크를 검토한다.

### 리스크 4: module path와 repo/binary 이름 혼선

대응:
- `go.mod`의 module path를 바꿀지 먼저 결정한다.
- 바꾸는 경우 문서/임포트/릴리즈 경로까지 같이 정리한다.

---

## 9. 권장 의사결정

내 판단으로는 지금 프로젝트에 가장 맞는 구조는 다음이다.

- 1 binary
- 1 TUI application
- 1 thin main
- 1 bootstrap layer
- 1 git adapter layer
- 1 graph domain layer
- 1 UI rendering layer
- `pkg/` 없음
- 서브커맨드 없음

즉, “클린 아키텍처 흉내”보다는 Go다운 단순한 레이어 분리가 맞다.

---

## 10. 문서 분리 제안

이 계획 문서는 현재 목적을 위해 충분하지만, 앞으로 구현이 시작되면 아래처럼 분리하는 것이 더 좋다.

### 분리 기준

1. 이 문서에는 "무엇을 어떻게 구조화할지"만 남긴다.
2. lazygit 분석처럼 외부 사례 비교는 별도 연구 문서로 뺀다.
3. 실제 구현 순서와 체크리스트는 별도 실행 문서로 유지한다.

### 권장 분리안

- `docs/cli-structure-research.md`
  - lazygit / 다른 Go CLI 구조 비교
  - 왜 `internal` 중심이 맞는지에 대한 근거
  - 패키지 설계 대안 비교
- `docs/cli-structure-implementation.md`
  - 실제 리팩토링 순서
  - 파일 이동/분리 작업
  - 검증 명령과 완료 기준

### 현재 문서의 역할

이 파일은 당장은 "최종 방향 + 마이그레이션 개요" 역할로 유지한다.
구현 직전에 더 촘촘한 실행 계획이 필요하면, 위의 implementation 문서를 따로 만들어서 이 문서와 분리하는 편이 좋다.

---

## 11. 다음 실행 순서

1. binary 이름과 module naming 원칙 확정
2. `cmd/graphkeeper` + `internal/bootstrap` 먼저 생성
3. `internal/graph`를 분리해서 순수 로직부터 이동
4. `internal/app/model.go`를 update/commands/navigation으로 쪼갬
5. view/theme 정리
6. 문서 갱신
7. 전체 테스트와 빌드 검증

---

## 12. 결론

`graphkeeper`는 지금 상태에서도 잘 출발했지만, 2026년 기준의 유지보수 가능한 Go CLI로 만들려면 다음이 중요하다.

- entrypoint는 얇게
- domain 로직은 순수하게
- UI는 표현만
- git 실행은 adapter로
- binary 이름과 프로젝트 이름을 일관되게

이 계획대로 가면 현재 MVP를 크게 흔들지 않으면서도, 이후 기능 확장과 리팩토링이 훨씬 쉬운 구조로 전환할 수 있다.
