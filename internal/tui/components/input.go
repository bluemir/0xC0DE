package components

import tea "charm.land/bubbletea/v2"

// Input 은 포커스된 행이 키를 위임받아 자기 상태를 제자리에서 갱신하는 위젯이다.
// 호스트가 포인터(예: &pawnInput.Sex)를 들고 Update 를 부르면 그 자리에서 바뀐다.
// 반환하는 tea.Cmd 는 Text 의 커서 깜빡임 같은 후속 작업용이다(Select/Number 는 nil).
type Input interface {
	Update(msg tea.Msg) tea.Cmd
	//Value() int
	Focus() tea.Cmd
	Blur()
}

// None 은 아무 행도 포커스되지 않은 상태를 나타내는 placeholder Input(키를 무시).
// 화면 진입 직후처럼 어떤 행도 선택되지 않았을 때 focus 자리에 둬, nil 가드 없이 Update 를 부를 수 있게 한다.
type none struct{}

func (none) Update(tea.Msg) tea.Cmd { return nil }
func (none) Focus() tea.Cmd         { return nil }
func (none) Blur()                  {}

var None Input = none{}

func GetValue[T ~float32 | ~int](i Input) T {
	switch input := i.(type) {
	case *Select:
		return T(input.Index())
	case *Number:
		return T(input.value)
	}
	// TODO
	return T(0)
}
