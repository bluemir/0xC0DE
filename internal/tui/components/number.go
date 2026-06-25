package components

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// Number 는 [min,max] 범위 안에서 step 만큼 ←→ 로 조정하는 수치 위젯이다(예: 나이·키·치수).
type Number struct {
	value, min, max, step int
	unit                  string
}

// NewNumber 는 value 에서 시작해 [min,max] 범위·step 단위·unit 표기를 가진 Number 를 만든다.
func NewNumber(value, min, max, step int, unit string) Number {
	// TODO functional parameter
	return Number{value: value, min: min, max: max, step: step, unit: unit}
}

// Dec/Inc 는 step 만큼 값을 옮긴다(범위에서 클램프).
func (n Number) Dec() Number {
	if n.value -= n.step; n.value < n.min {
		n.value = n.min
	}
	return n
}
func (n Number) Inc() Number {
	if n.value += n.step; n.value > n.max {
		n.value = n.max
	}
	return n
}

// Update 는 포커스됐을 때 ←→ 로 값을 옮긴다(제자리 갱신). 후속 명령은 없다(Input).
func (n *Number) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "right":
			*n = n.Inc()
		case "left":
			*n = n.Dec()
		}
	}
	return nil
}

func (n Number) Int() int       { return n.value }
func (n Number) String() string { return fmt.Sprintf("< %d%s >", n.value, n.unit) }
func (n Number) Focus() tea.Cmd { return nil }
func (n Number) Blur()          {}
