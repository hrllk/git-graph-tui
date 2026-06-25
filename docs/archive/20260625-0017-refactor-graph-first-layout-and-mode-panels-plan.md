# Refactor Graph-first Layout / Mode Panel Split Plan

## 목적

이 문서는 관리자 중심의 `graph-first` 레이아웃을 확정하고, `Mode` 섹션을 `Global` / `Context` 두 영역으로 분리하기 위한 구현 계획서다.

이번 계획의 핵심은 다음이다.

1. 그래프를 가장 넓은 작업면으로 배치한다.
2. `Local / Remote / Tags` 는 우측 보조 레일로 내린다.
3. `Mode` 는 전역 정보와 현재 포커스 정보로 분리한다.
4. top margin 은 유지한다.
5. `Graph / right rail` 은 3개의 동일 row height 로 나눈다.
6. `Graph` 는 왼쪽에서 3 rows 를 span 하고, `Local / Remote / Tags` 는 오른쪽에서 각각 1 row 씩 배치한다.
7. `graphPageSize` 는 outer margin 이 아니라 실제 `graph box height` 만 기준으로 계산한다.

## 참조 문서

이 계획은 다음 문서를 참고한다.

- `docs/archive/20260625-0015-feature-pull-reset-ux-implementation-plan.md`
- `docs/archive/20260625-0016-feature-reset-stash-plan.md`
- `docs/archive/20260625-0014-refactor-view-boundary-plan.md`
- `docs/archive/20260625-0012-refactor-navigation-boundary-plan.md`
- `docs/roadmap.md`
- `docs/structure.md`

## 현재 관찰된 구현 상태

현재 화면은 다음 구조로 배치되어 있다.

- 상단: `Local / Remote / Tags`
- 하단: `Graph / Detail`

이 구조는 정보 밀도는 나쁘지 않지만, 그래프가 주 작업면으로 보이지 않는다.

현재 코드에서는 다음 파일이 이 배치를 담당한다.

- `internal/app/view_shell.go`
- `internal/app/view_detail.go`
- `internal/app/view_sections.go`
- `internal/app/view_graph.go`
- `internal/graph/layout.go`

## 핵심 결정

### 1. 화면 구조는 4-row grid 로 본다

빨간 섹션 기준의 정렬은 다음처럼 구현한다.

```text
row 1: top margin
row 2: Global | Context
row 3: Graph   | Local
row 4: Graph   | Remote
row 5: Graph   | Tags
```

이때:

- `Global / Context` 는 top header row
- `Graph` 는 좌측에서 3 rows span
- `Local / Remote / Tags` 는 우측에서 각각 1 row 씩 차지

### 2. top margin 은 실제 빈 공간으로 유지한다

newline 누적이 아니라, 맨 위에 실제 빈 블록을 둔다.

이렇게 해야 천장에 붙어 보이지 않는다.

### 3. top header 높이는 작게 유지한다

상단 `Global / Context` 는 정보만 짧게 보여주는 영역이다.

권장 높이:

- `Global`: 현재 모드 + repo 상태 요약
- `Context`: 현재 섹션 이름 + 섹션별 핵심 액션

### 4. graphPageSize 는 graph box height 기준으로 계산한다

`PageSize` 는 outer margin, footer, 상단 header 전체를 알 필요가 없다.

`view_graph.go` 에서 실제 `graph box height` 를 계산하고, 그 값만 넘긴다.

권장 원칙:

- `graphPageSize = graphBoxHeight - paddingAllowance`
- outer margin 은 page size 계산에 포함하지 않는다

### 5. 우측 레일은 3개의 동일 row height 를 사용한다

우측 레일은 `Local / Remote / Tags` 를 동일한 row 규칙으로 쪼갠다.

허용되는 오차는 마지막 row 에만 흡수한다.

예시:

- `rowH := graphRailHeight / 3`
- `tagsH := graphRailHeight - localH - remoteH`

### 6. 작은 화면에서는 폴백이 필요하다

화면 폭이 충분하지 않으면 다음 순서로 축소한다.

1. 우측 레일 폭 축소
2. `Context` 상세 축약
3. `Tags` 우선 축약
4. 그래프 최소 폭 보장

그래프가 지나치게 좁아지면 의도한 UX 가 깨지므로, 그래프 최소 폭은 보장해야 한다.

## UI 배치 제안

### Global

`Global` 은 전역 정보만 보여준다.

권장 항목:

- 현재 모드
- repo 상태 요약
- 전역 hotkey
- 공통 입력 안내

### Context

`Context` 는 현재 포커스된 섹션 기준으로 보여준다.

권장 항목:

- 포커스된 섹션 이름
- 현재 선택된 항목 요약
- 해당 섹션에서 가능한 action
- 해당 섹션에서 확인해야 하는 상태

### Graph

Graph 는 가장 넓게 둔다.

권장 내용:

- commit
- branches
- graph lane
- when
- title

### Right rail

우측 레일은 세 개의 세로 블록으로 나눈다.

- `Local`
- `Remote`
- `Tags`

각 블록은 짧은 목록만 보여준다.

- `Local`: 현재 브랜치와 상태 마커
- `Remote`: origin / default 브랜치 요약
- `Tags`: 존재 여부와 짧은 목록

## 구현 순서

1. `view_shell.go` 를 4-row grid 형태로 바꾼다.
2. top margin 과 header height 를 분리한다.
3. `Global / Context` 를 header row 로 배치한다.
4. `Graph` 를 좌측 3 rows span 으로 배치한다.
5. `Local / Remote / Tags` 를 우측 3 rows 로 배치한다.
6. `graph.PageSize` 를 `graph box height` 기준으로 바꾼다.
7. 관련 렌더 테스트를 추가한다.

## Code Sketch

### `internal/app/view_shell.go`

```go
func renderAppView(m model) string {
	outerW := m.width
	outerH := m.height - 2
	if outerW < 80 {
		outerW = 80
	}
	if outerH < 18 {
		outerH = 18
	}

	hMargin := max(2, int(float64(m.width)*0.08))
	topMargin := max(1, int(float64(m.height)*0.08))
	bodyW := outerW - hMargin*2
	if bodyW < 80 {
		bodyW = 80
	}

	headerH := 4
	graphRailH := outerH - topMargin - headerH
	if graphRailH < 12 {
		graphRailH = 12
	}

	header := renderHeaderRow(m, bodyW, headerH)
	grid := renderGraphRailGrid(m, bodyW, graphRailH)

	body := strings.Join([]string{
		strings.Repeat("\n", topMargin),
		header,
		grid,
	}, "\n")

	body = applyHorizontalMargins(body, hMargin)
	return lipgloss.Place(outerW, outerH, lipgloss.Left, lipgloss.Top, body)
}
```

```go
func renderHeaderRow(m model, w, h int) string {
	leftW, rightW := splitPaneWidths(w)
	globalBox := baseBox.Width(leftW).Height(h).Render(
		"Global\n" + m.renderGlobalContent(max(leftW-4, 0), max(h-3, 0)),
	)
	contextBox := baseBox.Width(rightW).Height(h).Render(
		"Context\n" + m.renderContextContent(max(rightW-4, 0), max(h-3, 0)),
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, globalBox, contextBox)
}
```

```go
func renderGraphRailGrid(m model, w, h int) string {
	leftW := max(56, int(float64(w)*0.72))
	if leftW > w-18 {
		leftW = w - 18
	}
	if leftW < 0 {
		leftW = 0
	}
	rightW := w - leftW

	rowH := h / 3
	if rowH < 3 {
		rowH = 3
	}
	localH := rowH
	remoteH := rowH
	tagsH := h - localH - remoteH
	if tagsH < 3 {
		tagsH = 3
	}

	graphBox := m.getBoxStyle(sectionGraph).Width(leftW).Height(h).Render(
		"Graph\n" + m.renderGraphContent(max(leftW-4, 0), max(h-3, 0)),
	)
	localBox := m.getBoxStyle(sectionCurrent).Width(rightW).Height(localH).Render(
		"Local\n" + m.renderSectionContent(sectionCurrent, max(rightW-4, 0), max(localH-3, 0)),
	)
	remoteBox := m.getBoxStyle(sectionRemote).Width(rightW).Height(remoteH).Render(
		"Remote\n" + m.renderSectionContent(sectionRemote, max(rightW-4, 0), max(remoteH-3, 0)),
	)
	tagsBox := m.getBoxStyle(sectionTags).Width(rightW).Height(tagsH).Render(
		"Tags\n" + m.renderSectionContent(sectionTags, max(rightW-4, 0), max(tagsH-3, 0)),
	)

	rightRail := lipgloss.JoinVertical(lipgloss.Left, localBox, remoteBox, tagsBox)
	return lipgloss.JoinHorizontal(lipgloss.Top, graphBox, rightRail)
}
```

### `internal/graph/layout.go`

```go
func PageSize(graphBoxHeight int) int {
	if graphBoxHeight <= 0 {
		return 3
	}
	page := graphBoxHeight - 2
	if page < 3 {
		page = 3
	}
	return page
}
```

### `internal/app/view_graph.go`

```go
func (m model) renderGraphContent(width, height int) string {
	graphBoxHeight := height
	page := graph.PageSize(graphBoxHeight)
	start := clampScroll(m.graphScroll, len(rows), page)
	end := min(len(rows), start+page)
	...
}
```

### `internal/app/view_detail.go`

```go
func (m model) renderGlobalContent(width, height int) string {
	lines := []string{
		title.Render("Mode"),
		renderStatusCompact(m.status),
		"",
		title.Render("Hotkeys"),
		"1 local  •  2 remote  •  3 tags  •  4 graph",
		"tab/shift+tab section  •  up/down/j/k move",
	}
	return fitBlockLines(lines, height)
}

func (m model) renderContextContent(width, height int) string {
	lines := []string{
		title.Render("Context"),
		sectionName(m.activeSection),
	}
	lines = append(lines, renderSectionContextLines(m, width)...)
	return fitBlockLines(lines, height)
}
```

### `internal/app/model.go`

```go
type graphSection int

const (
	sectionGraph graphSection = iota
	sectionCurrent
	sectionRemote
	sectionTags
)
```

## 테스트 계획

```go
func TestRenderAppViewUsesTopMargin(t *testing.T)
func TestRenderAppViewUsesGraphRailGrid(t *testing.T)
func TestGraphPageSizeUsesGraphBoxHeight(t *testing.T)
func TestRightRailRowsSplitEvenly(t *testing.T)
func TestGraphRowSpanRemainsAlignedWithTags(t *testing.T)
```

## 테스트 항목

- top margin 이 실제 빈 공간으로 보이는지
- `Global / Context` 가 header row 에만 보이는지
- `Graph` 가 좌측에서 3 rows span 하는지
- `Local / Remote / Tags` 가 우측에서 1 row 씩 보이는지
- `graphPageSize` 가 outer margin 에 영향을 받지 않는지
- `Tags` 높이가 `Graph` 하단 정렬과 어긋나지 않는지

## 내일 이어서 할 일

- [ ] `view_shell.go` 를 4-row grid 로 교체한다
- [ ] `graph.PageSize` 입력값을 `graphBoxHeight` 로 바꾼다
- [ ] header / rail row 높이 테스트를 추가한다
- [ ] 실제 스크린샷으로 row span 을 다시 확인한다

## 결론

이 레이아웃은 단순 비율 조정이 아니라, 화면을 명시적인 row span 구조로 다시 정의하는 작업이다.

- top margin 은 분리
- header 는 작게
- graph 는 3 rows span
- right rail 은 3 rows
- page size 는 실제 graph box height 기준

이 기준으로 맞추면 빨간 섹션대로 구현 가능하다.
