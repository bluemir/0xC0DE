package components

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// Text 는 한 줄 텍스트 입력 위젯이다. bubbles 의 textinput.Model 을 감싸 다른 위젯과 같은 Input 으로
// 통일한다(포커스된 행이 키를 위임받아 제자리에서 갱신). textinput 은 포커스 상태에서만 키를 처리하므로
// 생성 시 Focus 해 둔다.
type Text struct {
	model textinput.Model
}

// NewText 는 초기값 value 로 포커스된 텍스트 입력을 만든다.
func NewText(value string) Text {
	ti := textinput.New()
	ti.SetValue(value)
	ti.Focus()
	return Text{model: ti}
}

// Update 는 타이핑·캐럿 이동·커서 깜빡임을 textinput 에 위임한다(제자리 갱신). 깜빡임을 이어가는 후속
// 명령이 있으면 그 tea.Cmd 를 돌려준다.
func (t *Text) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return cmd
}

// Focus/Blur 는 편집(캐럿·키 수신·깜빡임) 활성 여부를 토글한다. 호스트가 이 입력을 현재 행으로 들이고
// 낼 때 부른다. Focus 는 깜빡임을 시작하는 tea.Cmd 를 돌려준다.
func (t *Text) Focus() tea.Cmd { return t.model.Focus() }
func (t *Text) Blur()          { t.model.Blur() }

func (t Text) Value() string  { return t.model.Value() }
func (t Text) String() string { return t.model.View() }

// Blink 는 텍스트 커서 깜빡임을 시작하는 명령이다(Text 를 쓰는 화면의 Init/reroll 에서 반환).
func Blink() tea.Msg { return textinput.Blink() }
