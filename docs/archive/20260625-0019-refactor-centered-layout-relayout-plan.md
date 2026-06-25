# Centered Graph Layout / Style Relayout Plan

> **For Hermes:** Use this plan task-by-task if you split implementation work.

**Goal:** 현재 레이아웃을 `가운데 정렬 + 10% 외곽 여백` 기준으로 다시 잡고, `Global / Context / Graph / Local / Remote / Tags` 의 시각적 비율과 단축키 배치를 현재 UX 요구사항에 맞게 정리한다.

**Architecture:** `internal/app` 의 렌더링, 레이아웃 계산, 키 처리, 팝업 오버레이를 함께 손본다. 화면 구조는 유지하되, 배치 규칙과 계산식은 정리한다. 새 패키지는 만들지 않는다.

**Tech Stack:** Go, Bubble Tea (`tea.Model`, `tea.KeyMsg`), Lip Gloss, `go test`

---

### TOC

- ### Goal
- ### Scope
- ### Review Notes
- ### BEFORE
- ### AFTER
- ### Tests
- ### Verification
- ### Notes

---

### Goal

이번 작업의 핵심은 단순한 간격 조정이 아니라, 화면 전체의 기준점을 다시 잡는 것이다.

현재 화면은 기능적으로는 동작하지만 다음 문제가 겹쳐 있다.

- 전체 레이아웃이 좌상단 기준으로 보이고, 제품 화면답게 가운데 정렬되지 않는다.
- 외곽 여백이 충분히 보이지 않아 천장과 벽에 붙은 느낌이 난다.
- `Global / Context` 비율이 1:1로 보여서 정보 위계가 흐릿하다.
- `Graph` 섹션의 노출량이 체감상 부족하고, 실제 사용 가능한 높이와 렌더링 높이가 어긋난다.
- `Graph` 와 우측 `Local / Remote / Tags` 의 세로 기준이 정확히 맞아야 한다.
- 숫자 키 `1/2/3/4` 의 섹션 매핑이 현재 요구사항과 반대로 되어 있다.
- confirm / reset-mode popup 이 전체 레이아웃과 같은 좌표계를 공유하면서 스타일이 깨질 수 있다.

요구사항을 한 문장으로 정리하면 다음과 같다.

- 화면은 가운데 정렬
- 전체 마진은 약 10%
- 상단 `Global / Context` 는 3:7 비율
- 섹션 키는 `1 Graph, 2 Local, 3 Remote, 4 Tags`
- `1/2/3/4` 외의 global hotkey 는 화면의 `Global` 섹션으로 이동
- `Graph` 높이는 우측 `Local + Remote + Tags` 높이 합과 정확히 일치
- popup 은 레이아웃을 깨지 않도록 별도 오버레이 규칙을 사용

---

### Scope

- `internal/app/view_shell.go`
- `internal/app/view_layout.go`
- `internal/app/view_detail.go`
- `internal/app/key_handling_browse.go`
- `internal/app/navigation_section.go`
- `internal/app/model_test.go`
- 필요하면 `internal/app/key_handling.go` 와 `internal/app/view_sections.go` 의 안내 문구도 함께 맞춘다

이번 범위에서는 다음을 지킨다.

- 레이아웃 계산식은 정리하되, 앱의 정보 구조는 크게 바꾸지 않는다.
- `Global`, `Context`, `Graph`, `Local`, `Remote`, `Tags` 섹션 자체는 유지한다.
- `1/2/3/4` 숫자 키의 의미만 새 요구사항에 맞게 바꾼다.
- popup 은 별도 시스템으로 분리하지 않고, 현재 렌더링 계층 안에서 안정화한다.

---

### Review Notes

- 현재 `renderAppView` 는 `headerRow` 와 `graphRow` 를 같은 본문에 넣고 `applyOuterMargins` 로 감싼다. 이 구조는 유지 가능하지만, 외곽 여백 계산과 가운데 배치 기준을 다시 정의해야 한다.
- `layoutShellMargins` 는 현재 8% / 4% 기준이다. 요구사항은 10% 수준의 외곽 마진과 더 명확한 중앙 정렬이다.
- `splitPaneWidths` 는 현재 50:50 성격으로 쓰이고 있어 `Global / Context` 위계가 약하다.
- `renderGlobalContent` 의 hotkey 안내는 현재 `1 local • 2 remote • 3 tags • 4 graph` 로 되어 있다. 이는 새 섹션 이동 규칙과 정면으로 충돌한다.
- `handleBrowseGlobalKey` 는 현재 `1=Local, 2=Remote, 3=Tags, 4=Graph` 로 동작한다. 이것을 `1=Graph, 2=Local, 3=Remote, 4=Tags` 로 바꿔야 한다.
- `overlayPopup` 는 기본적으로 문자열 위에 popup 을 덮는 방식이라, popup 폭과 base 폭이 엇갈리면 시각적 깨짐이 발생할 수 있다.
- `graphPageSize` 는 현재 `graphBoxHeightForModel` 에 기반한다. 이 계산은 유지하되, 실제 body height 와 중복 차감이 생기지 않도록 기준을 명확히 해야 한다.

---

### BEFORE

현재 구조는 대략 다음과 같다.

```go
func renderAppView(m model) string {
	hMargin, topMargin, bottomMargin := layoutShellMargins(m)
	bodyWidth, bodyHeight := layoutShellBodySize(m, hMargin, topMargin, bottomMargin)
	headerHeight := layoutHeaderHeight(bodyHeight)
	graphRailHeight := layoutGraphRailHeight(bodyHeight)

	globalWidth, contextWidth := splitPaneWidths(bodyWidth)
	globalBox := baseBox.Width(globalWidth).Height(headerHeight).Render(
		"Global\n" + m.renderGlobalContent(max(globalWidth-4, 0), max(headerHeight-3, 0)),
	)
	contextBox := baseBox.Width(contextWidth).Height(headerHeight).Render(
		"Context\n" + m.renderContextContent(max(contextWidth-4, 0), max(headerHeight-3, 0)),
	)
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, globalBox, contextBox)

	graphWidth := int(float64(bodyWidth) * 0.72)
	if graphWidth < 56 {
		graphWidth = 56
	}
	if graphWidth > bodyWidth-18 {
		graphWidth = bodyWidth - 18
	}
	if graphWidth < 0 {
		graphWidth = 0
	}
	rightWidth := bodyWidth - graphWidth
	graphBox := m.getBoxStyle(sectionGraph).Width(graphWidth).Height(graphRailHeight).Render(
		"Graph\n" + m.renderGraphContent(max(graphWidth-4, 0), max(graphRailHeight-3, 0)),
	)
	rightRail := m.renderRightRail(rightWidth, graphRailHeight)
	graphRow := lipgloss.JoinHorizontal(lipgloss.Top, graphBox, rightRail)

	body := lipgloss.JoinVertical(lipgloss.Left, headerRow, graphRow)
	body = lipgloss.Place(bodyWidth, bodyHeight, lipgloss.Left, lipgloss.Top, body)
	centeredBody := applyOuterMargins(body, bodyWidth, bodyHeight, hMargin, topMargin, bottomMargin)
	// ...
}
```

```go
func (m model) renderGlobalContent(width, height int) string {
	lines := append(lines, title.Render("Hotkeys"))
	lines = append(lines, "1 local  •  2 remote  •  3 tags  •  4 graph")
	lines = append(lines, "tab/shift+tab section  •  up/down/j/k move")
	lines = append(lines, "f fetch  •  q quit")
}
```

```go
func (m model) handleBrowseGlobalKey(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1":
		m = switchBrowseSection(m, sectionCurrent)
	case "2":
		m = switchBrowseSection(m, sectionRemote)
	case "3":
		m = switchBrowseSection(m, sectionTags)
	case "4":
		m = switchBrowseSection(m, sectionGraph)
	}
}
```

```go
func overlayPopup(base string, popup string) string {
	baseLines := strings.Split(base, "\n")
	popupLines := strings.Split(popup, "\n")
	baseH := len(baseLines)
	popupH := len(popupLines)
	if baseH < popupH {
		return base
	}
	popupW := 0
	for _, l := range popupLines {
		w := lipgloss.Width(l)
		if w > popupW {
			popupW = w
		}
	}
	startY := (baseH - popupH) / 2
	// ...
}
```

이 구조는 동작은 하지만 다음 문제가 있다.

- `Global / Context` 폭이 같아서 상단 정보 위계가 약하다.
- 숫자 키 매핑이 현재 요구사항과 반대다.
- `renderGlobalContent` 에서 안내하는 단축키와 실제 동작이 어긋난다.
- popup 은 body 문자열 위에 직접 합성되어, 폭/높이 클램프가 없으면 스타일이 쉽게 무너진다.

---

### AFTER

#### 1) 중앙 정렬과 10% 외곽 여백을 레이아웃 기준으로 고정한다

레이아웃 계산은 `10% safe frame` 을 기준으로 하고, 최종 프레임은 화면 가운데에 배치한다.

```go
func layoutShellMargins(m model) (hMargin, topMargin, bottomMargin int) {
	hMargin = int(float64(m.width) * 0.10)
	topMargin = int(float64(m.height) * 0.10)
	bottomMargin = int(float64(m.height) * 0.10)

	if hMargin < 2 {
		hMargin = 2
	}
	if topMargin < 1 {
		topMargin = 1
	}
	if bottomMargin < 1 {
		bottomMargin = 1
	}

	maxSide := (m.width - 80) / 2
	if maxSide >= 0 && hMargin > maxSide {
		hMargin = maxSide
	}

	maxVertical := m.height - 20
	if maxVertical >= 0 && topMargin > maxVertical {
		topMargin = maxVertical
	}
	if maxBottom := m.height - topMargin - 19; maxBottom >= 0 && bottomMargin > maxBottom {
		bottomMargin = maxBottom
	}
	return hMargin, topMargin, bottomMargin
}

func layoutShellBodySize(m model, hMargin, topMargin, bottomMargin int) (width, height int) {
	width = m.width - hMargin*2
	if width < 80 {
		width = 80
	}

	height = m.height - topMargin - bottomMargin - 1
	if height < 18 {
		height = 18
	}
	return width, height
}
```

#### 2) `Global / Context` 는 3:7 비율로 분리한다

상단은 전역 정보와 컨텍스트를 명확히 나눈다.

```go
func splitPaneWidths(total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}

	left := total * 3 / 10
	if left < 28 {
		left = 28
	}
	if left > total-32 {
		left = total - 32
	}
	if left < 0 {
		left = 0
	}
	right := total - left
	return left, right
}

func renderAppView(m model) string {
	hMargin, topMargin, bottomMargin := layoutShellMargins(m)
	bodyWidth, bodyHeight := layoutShellBodySize(m, hMargin, topMargin, bottomMargin)
	headerHeight := layoutHeaderHeight(bodyHeight)
	graphRailHeight := layoutGraphRailHeight(bodyHeight)

	globalWidth, contextWidth := splitPaneWidths(bodyWidth)
	globalBox := baseBox.Width(globalWidth).Height(headerHeight).Render(
		"Global\n" + m.renderGlobalContent(max(globalWidth-4, 0), max(headerHeight-3, 0)),
	)
	contextBox := baseBox.Width(contextWidth).Height(headerHeight).Render(
		"Context\n" + m.renderContextContent(max(contextWidth-4, 0), max(headerHeight-3, 0)),
	)
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, globalBox, contextBox)

	graphWidth := bodyWidth * 72 / 100
	if graphWidth < 56 {
		graphWidth = 56
	}
	if graphWidth > bodyWidth-18 {
		graphWidth = bodyWidth - 18
	}
	if graphWidth < 0 {
		graphWidth = 0
	}
	rightWidth := bodyWidth - graphWidth

	graphBox := m.getBoxStyle(sectionGraph).Width(graphWidth).Height(graphRailHeight).Render(
		"Graph\n" + m.renderGraphContent(max(graphWidth-4, 0), max(graphRailHeight-3, 0)),
	)
	rightRail := m.renderRightRail(rightWidth, graphRailHeight)
	graphRow := lipgloss.JoinHorizontal(lipgloss.Top, graphBox, rightRail)

	body := lipgloss.JoinVertical(lipgloss.Left, headerRow, graphRow)
	frame := applyOuterMargins(body, bodyWidth, bodyHeight, hMargin, topMargin, bottomMargin)
	centeredBody := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, frame)

	if m.status.Mode == state.ModeConfirm || m.status.Mode == state.ModeResetModePick {
		centeredBody = overlayPopup(centeredBody, renderModalPopup(m, bodyWidth, bodyHeight))
	}

	footer := muted.Render("tab/shift+tab section  •  up/down/j/k move  •  f fetch  •  q quit")
	footer = lipgloss.Place(bodyWidth, 1, lipgloss.Center, lipgloss.Center, footer)
	footer = applyOuterMarginLine(footer, bodyWidth, hMargin)

	return centeredBody + strings.Repeat("\n", bottomMargin) + "\n" + footer + "\n"
}
```

#### 3) `Graph` 높이는 우측 세 섹션 합과 정확히 맞춘다

우측 레일은 항상 같은 기준 높이를 사용하고, `Graph` 는 그 합을 그대로 점유한다.

```go
func splitThreeHeights(total int) (int, int, int) {
	if total <= 0 {
		return 0, 0, 0
	}

	local := total / 3
	remote := total / 3
	tags := total - local - remote

	if local == 0 {
		local = 1
	}
	if remote == 0 && total > 1 {
		remote = 1
	}
	if tags == 0 && total > 2 {
		tags = 1
	}

	for local+remote+tags > total {
		switch {
		case tags > 1:
			tags--
		case remote > 1:
			remote--
		case local > 1:
			local--
		default:
			return total, 0, 0
		}
	}

	if rem := total - (local + remote + tags); rem > 0 {
		tags += rem
	}
	return local, remote, tags
}

func (m model) renderRightRail(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	localHeight, remoteHeight, tagsHeight := splitThreeHeights(height)

	localBox := m.getBoxStyle(sectionCurrent).Width(width).Height(localHeight).Render(
		"Local\n" + m.renderSectionContent(sectionCurrent, max(width-4, 0), max(localHeight-3, 0)),
	)
	remoteBox := m.getBoxStyle(sectionRemote).Width(width).Height(remoteHeight).Render(
		"Remote\n" + m.renderSectionContent(sectionRemote, max(width-4, 0), max(remoteHeight-3, 0)),
	)
	tagsBox := m.getBoxStyle(sectionTags).Width(width).Height(tagsHeight).Render(
		"Tags\n" + m.renderSectionContent(sectionTags, max(width-4, 0), max(tagsHeight-3, 0)),
	)

	return lipgloss.JoinVertical(lipgloss.Left, localBox, remoteBox, tagsBox)
}
```

`Graph` 쪽 페이지 계산은 그대로 유지하되, 아래 규칙을 명시적으로 지킨다.

```go
func graphPageSize(m *model) int {
	return graph.PageSize(graphBoxHeightForModel(m))
}

func graphBoxHeightForModel(m *model) int {
	hMargin, topMargin, bottomMargin := layoutShellMargins(*m)
	_, bodyHeight := layoutShellBodySize(*m, hMargin, topMargin, bottomMargin)
	return layoutGraphRailHeight(bodyHeight)
}
```

#### 4) 섹션 이동키를 새 순서로 재배치한다

숫자 키는 `1 Graph, 2 Local, 3 Remote, 4 Tags` 로 바뀐다.

```go
func (m model) handleBrowseGlobalKey(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return true, m, tea.Quit
	case "1":
		m = switchBrowseSection(m, sectionGraph)
		return true, m, nil
	case "2":
		m = switchBrowseSection(m, sectionCurrent)
		return true, m, nil
	case "3":
		m = switchBrowseSection(m, sectionRemote)
		return true, m, nil
	case "4":
		m = switchBrowseSection(m, sectionTags)
		return true, m, nil
	// 나머지 키는 기존 흐름 유지
	}
	return false, m, nil
}
```

섹션 순서도 이 규칙에 맞춰 고정한다.

```go
func graphSectionOrder() []graphSection {
	return []graphSection{sectionGraph, sectionCurrent, sectionRemote, sectionTags}
}

func sectionName(section graphSection) string {
	switch section {
	case sectionGraph:
		return "Graph"
	case sectionCurrent:
		return "Local"
	case sectionRemote:
		return "Remote"
	case sectionTags:
		return "Tags"
	default:
		return "Unknown"
	}
}
```

`tab / shift+tab` 도 이 순서를 따르며, `sectionName` 은 화면에 보이는 이름을 `Local` 로 통일한다.

#### 5) `Global` 섹션으로 숫자키 외의 hotkey 안내를 이동한다

Global 패널은 섹션 이동키 안내를 제거하고, 공통 동작만 보여준다.

```go
func (m model) renderGlobalContent(width, height int) string {
	if height <= 0 {
		return ""
	}

	lines := make([]string, 0, height)
	lines = append(lines, title.Render("Mode"))
	lines = append(lines, renderStatusCompact(m.status))
	lines = append(lines, "")

	lines = append(lines, title.Render("Repo"))
	lines = append(lines, fmt.Sprintf("branch: %-12s • head: %s", shorten(m.repoStatus.Branch, 10), shorten(m.repoStatus.Head, 7)))
	lines = append(lines, fmt.Sprintf("upstream: %-10s • remote: %s", shorten(emptyDash(m.repoStatus.Upstream), 10), shorten(emptyDash(m.repoStatus.Remote), 10)))

	lines = append(lines, "")
	lines = append(lines, title.Render("Hotkeys"))
	lines = append(lines, "1 graph  •  2 local  •  3 remote  •  4 tags")
	lines = append(lines, "tab/shift+tab section  •  up/down/j/k move")
	lines = append(lines, "f fetch  •  q quit")
	return fitBlockLines(lines, height)
}
```

`footer` 는 중복 안내를 줄여서, 숫자키가 두 번 보이지 않도록 정리한다.

#### 6) popup 은 별도 폭 클램프와 중앙 배치 규칙을 가진다

popup 은 body 문자열에 그냥 덮지 말고, 실제 terminal size 에 맞춰 먼저 안전하게 만든다.

```go
func renderModalPopup(m model, bodyWidth, bodyHeight int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	popupWidth := bodyWidth - 12
	if popupWidth > 54 {
		popupWidth = 54
	}
	if popupWidth < 32 {
		popupWidth = 32
	}

	popupBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(popupWidth).
		Align(lipgloss.Center)

	popupTitle := m.status.Title
	if popupTitle == "" || popupTitle == "Confirm" {
		popupTitle = "Continue?"
	}

	helpText := "y: yes  •  n: no"
	if m.status.Action == state.ActionPull && !m.pullIsFastForward {
		helpText = "m: merge  •  r: rebase  •  esc: cancel"
	} else if m.status.Mode == state.ModeResetModePick {
		helpText = "s: soft  •  m: mixed  •  h: hard  •  enter: execute  •  esc: back"
	}

	return popupBox.Render(
		titleStyle.Render(popupTitle) + "\n\n" +
			descStyle.Render(m.status.Detail) + "\n\n" +
			helpStyle.Render(helpText),
	)
}
```

오버레이는 popup 이 body 보다 커지면 자동으로 잘라내거나 축소되도록 유지한다.

```go
func overlayPopup(base string, popup string) string {
	baseLines := strings.Split(base, "\n")
	popupLines := strings.Split(popup, "\n")
	baseH := len(baseLines)
	popupH := len(popupLines)
	if baseH < popupH {
		return base
	}

	popupW := 0
	for _, line := range popupLines {
		if w := lipgloss.Width(line); w > popupW {
			popupW = w
		}
	}

	startY := (baseH - popupH) / 2
	for i, popupLine := range popupLines {
		y := startY + i
		if y >= len(baseLines) {
			break
		}
		baseLine := baseLines[y]
		baseLines[y] = overlayLine(baseLine, popupLine, max((lipgloss.Width(baseLine)-popupW)/2, 0), popupW)
	}
	return strings.Join(baseLines, "\n")
}
```

#### 7) 이동 안내와 스타일 테스트를 같이 고정한다

레이아웃은 눈으로만 보면 다시 흔들리기 쉽기 때문에, test 로 규칙을 박아 둔다.

---

### Tests

다음 케이스를 우선 고정한다.

```go
func TestShellLayoutCentersWithTenPercentMargins(t *testing.T) { /* ... */ }
func TestRenderAppViewUsesThreeSevenTopHeaderSplit(t *testing.T) { /* ... */ }
func TestRenderAppViewKeepsGraphHeightEqualToRailHeight(t *testing.T) { /* ... */ }
func TestRenderRightRailStacksLocalRemoteTagsInOrder(t *testing.T) { /* ... */ }
func TestGlobalHotkeyDigitsAreMovedIntoGlobalPanel(t *testing.T) { /* ... */ }
func TestHandleBrowseGlobalKeyMapsDigitsToGraphLocalRemoteTags(t *testing.T) { /* ... */ }
func TestSectionNamesShowLocalNotBranches(t *testing.T) { /* ... */ }
func TestPopupOverlayDoesNotBreakCenteredLayout(t *testing.T) { /* ... */ }
func TestFooterDoesNotRepeatSectionDigits(t *testing.T) { /* ... */ }
```

검증 포인트는 다음과 같다.

- 본문 시작 위치가 좌상단에 붙지 않는다.
- top margin 이 존재한다.
- `Global / Context` 가 3:7 으로 나뉜다.
- `Graph` 는 우측 레일 높이 합과 동일한 세로 영역을 갖는다.
- `Local / Remote / Tags` 는 항상 같은 순서로 쌓인다.
- 숫자키가 새 섹션 순서대로 동작한다.
- popup 표시 시 본문 border / padding / alignment 가 무너지지 않는다.

---

### Verification

구현 후에는 아래 순서로 확인한다.

1. `scripts/check`
2. `scripts/test`
3. 레이아웃 관련 렌더 테스트만 빠르게 돌려서 실패 지점을 좁힌다.
4. 실제 터미널에서 최소 한 번 확인한다.
5. confirm / reset mode / branch prompt 상황을 모두 열어 본다.

권장 추가 확인 항목:

- 80x24
- 120x40
- 140x60
- popup 열림
- popup 취소 후 레이아웃 복구
- 숫자키 `1/2/3/4`
- `tab / shift+tab`
- `up/down/j/k`

---

### Notes

- 이 문서는 `docs/archive/202606-0002-refactor-commands-test-plan.md` 의 문서 구조를 참고해서 작성했지만, 실제 내용은 레이아웃 리레이아웃용으로 바꿨다.
- 레이아웃 비율은 화면 크기에 따라 약간 달라질 수 있으나, 사용자가 체감하는 위계는 변하지 않아야 한다.
- `Local` 이라는 표기는 내부적으로는 `sectionCurrent` 를 쓰더라도 화면 표기에서는 유지한다.
- 만약 이후 `Global` 섹션에 항목이 더 늘어나면, footer 를 다시 늘리지 말고 `Global` 패널 내부에서만 정리한다.
- 이 계획이 적용되면 `docs/decisions.md` 에 "centered shell / 10% margin / section digits" 관련 결정을 추가하는 편이 좋다.
