// Package component 는 자기 상태를 들고 ←→ 로 조정되는 작은 재사용 TUI 입력 위젯을 모은다.
// Prev/Next·Inc/Dec 는 값 리시버 + 복사본 반환이지만, 포커스된 행이 키를 위임할 때 쓰는 Update 는
// 포인터 리시버라 호스트가 들고 있는 필드를 제자리에서 갱신한다(Input 인터페이스).
package components

import tea "charm.land/bubbletea/v2"

// NewSelect 는 options 중 idx 번째를 가리키는 Select 를 만든다.
func NewSelect(options []string, idx int) Select {
	return Select{options: options, idx: idx}
}

// Select 는 정해진 목록을 ←→ 로 순환 선택하는 위젯이다(예: 머리·성별).
type Select struct {
	options []string
	idx     int
	focused bool
}

var _ Input = (*Select)(nil)

// Prev/Next 는 선택을 한 칸 옮긴다(양 끝에서 순환).
func (s Select) Prev() Select {
	s.idx = (s.idx - 1 + len(s.options)) % len(s.options)
	return s
}
func (s Select) Next() Select {
	s.idx = (s.idx + 1) % len(s.options)
	return s
}

// Update 는 포커스됐을 때 ←→ 로 선택을 옮긴다(제자리 갱신). 후속 명령은 없다(Input).
func (s *Select) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "right":
			*s = s.Next()
		case "left":
			*s = s.Prev()
		}
	}
	return nil
}

func (s Select) Value() string  { return s.options[s.idx] }
func (s Select) Index() int     { return s.idx }
func (s Select) String() string { return renderArrows(s.options[s.idx], s.focused) }

// Focus/Blur 는 포커스 상태를 토글한다(화살표 색에 반영). 호스트가 이 입력을 현재 행으로 들이고 낼 때 부른다.
func (s *Select) Focus() tea.Cmd { s.focused = true; return nil }
func (s *Select) Blur()          { s.focused = false }
