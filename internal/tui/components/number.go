package components

import (
	"fmt"
	"math"

	tea "charm.land/bubbletea/v2"
)

// NewNumber 는 [min,max] 범위·step 단위·unit 표기를 가진 Number 를 만든다.
// 옵션을 주지 않으면 [math.MinInt, math.MaxInt] 범위에 step 1, unit 없음으로 시작한다.
func NewNumber(opts ...NumberOptionFn) Number {
	opt := NumberOption{
		// Defaults
		min:  math.MinInt,
		max:  math.MaxInt,
		step: 1,
		unit: "",
	}

	for _, fn := range opts {
		fn(&opt)
	}
	return NewNumberWithOption(opt)
}
func NewNumberWithOption(opt NumberOption) Number {
	value := opt.initial
	if value < opt.min {
		value = opt.min
	}
	if value > opt.max {
		value = opt.max
	}
	return Number{
		value: value,
		min:   opt.min,
		max:   opt.max,
		step:  opt.step,
		unit:  opt.unit,
	}
}

// Number 는 [min,max] 범위 안에서 step 만큼 ←→ 로 조정하는 수치 위젯이다(예: 나이·키·치수).
type Number struct {
	value, min, max, step int
	unit                  string
	focused               bool
}

type NumberOption struct {
	initial  int
	min, max int
	step     int
	unit     string
}
type NumberOptionFn func(*NumberOption)

func WithRange(min, max int) NumberOptionFn {
	return func(opt *NumberOption) {
		opt.max = max
		opt.min = min
	}
}
func WithStep(step int) NumberOptionFn {
	return func(opt *NumberOption) {
		opt.step = step
	}
}
func WithUnit(unit string) NumberOptionFn {
	return func(opt *NumberOption) {
		opt.unit = unit
	}
}
func WithInitial(initial int) NumberOptionFn {
	return func(opt *NumberOption) {
		opt.initial = initial
	}
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
func (n Number) String() string { return renderArrows(fmt.Sprintf("%d%s", n.value, n.unit), n.focused) }

// Focus/Blur 는 포커스 상태를 토글한다(화살표 색에 반영). 호스트가 이 입력을 현재 행으로 들이고 낼 때 부른다.
func (n *Number) Focus() tea.Cmd { n.focused = true; return nil }
func (n *Number) Blur()          { n.focused = false }
