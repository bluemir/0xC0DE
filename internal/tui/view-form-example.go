package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/bluemir/0xC0DE/internal/tui/components"
)

func FormExample() (tea.Model, tea.Cmd) {
	inputs := formInputs{
		Name:   components.NewText(""),
		Age:    components.NewNumber(0, 0, 100, 1, ""),
		Gender: components.NewSelect([]string{"male", "female"}, 0),
		Habby:  components.NewText(""),
	}
	return formSummary(&inputs)
}

type formInputs struct {
	Name   components.Text
	Age    components.Number
	Gender components.Select

	Habby components.Text
}

func (inputs *formInputs) blur() {
	inputs.Name.Blur()
	inputs.Age.Blur()
	inputs.Gender.Blur()
	inputs.Habby.Blur()
}

func formSummary(inputs *formInputs) (tea.Model, tea.Cmd) {
	inputs.blur()
	return &viewFormExampleSummary{inputs: inputs}, nil
}

type viewFormExampleSummary struct {
	inputs *formInputs
}

func (v viewFormExampleSummary) Init() tea.Cmd { return nil }
func (v viewFormExampleSummary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return QuitConfirm(v)
		case "tab":
			return formTab1(v.inputs)
		//case "shift+tab":
		default:
			return v, nil
		}
	default:
		return v, nil
	}
}
func (v viewFormExampleSummary) View() tea.View {
	return tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		"[>summary<] [ tab1 ] [ tab2 ]",
		//
		fmt.Sprintf("이름: %s", v.inputs.Name.Value()),
		fmt.Sprintf("나이: %d", v.inputs.Age.Int()),
		fmt.Sprintf("성별: %s", v.inputs.Gender.Value()),
		fmt.Sprintf("취미: %s", v.inputs.Habby.Value()),
	))
}

func formTab1(inputs *formInputs) (tea.Model, tea.Cmd) {
	inputs.blur()
	return &viewFormExampleTab1{inputs: inputs, focus: &inputs.Name}, inputs.Name.Focus()
}

type viewFormExampleTab1 struct {
	inputs *formInputs

	focus components.Input
}

func (v *viewFormExampleTab1) Init() tea.Cmd { return nil }

// Update 는 포인터 리시버다. 값 리시버면 호출마다 v.inputs 가 새 주소로 복사돼
// focus 포인터 비교(&v.inputs.Name 등)가 어긋나 포커스 이동이 동작하지 않는다.
func (v *viewFormExampleTab1) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return QuitConfirm(v)
		case "tab":
			return formTab2(v.inputs)
		case "shift+tab":
			return formSummary(v.inputs)
		case "up":
			switch v.focus {
			case &v.inputs.Name:
				//noop. start of form
			case &v.inputs.Age:
				v.focus.Blur()
				v.focus = &v.inputs.Name
				return v, v.focus.Focus()
			case &v.inputs.Gender:
				v.focus.Blur()
				v.focus = &v.inputs.Age
				return v, v.focus.Focus()
			}
			return v, nil
		case "down":
			switch v.focus {
			case &v.inputs.Name:
				v.focus.Blur()
				v.focus = &v.inputs.Age
				return v, v.focus.Focus()
			case &v.inputs.Age:
				v.focus.Blur()
				v.focus = &v.inputs.Gender
				return v, v.focus.Focus()
			case &v.inputs.Gender:
				// noop end
			}
			return v, nil
		default:
			// 타이핑·←→ 등 행 이동이 아닌 키는 포커스된 입력에 위임한다.
			return v, v.focus.Update(msg)
		}
	default:
		// 커서 깜빡임 등 키 외 메시지도 포커스된 입력에 위임한다.
		return v, v.focus.Update(msg)
	}
}
func (v *viewFormExampleTab1) View() tea.View {
	return tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		"[ summary ] [>tab1<] [ tab2 ]",
		//
		v.inputs.Name.String(),
		v.inputs.Age.String(),
		v.inputs.Gender.String(),
	))
}

func formTab2(inputs *formInputs) (tea.Model, tea.Cmd) {
	inputs.blur()

	return &viewFormExampleTab2{inputs: inputs, focus: &inputs.Habby}, inputs.Habby.Focus()
}

type viewFormExampleTab2 struct {
	inputs *formInputs

	focus components.Input
}

func (v *viewFormExampleTab2) Init() tea.Cmd { return nil }
func (v *viewFormExampleTab2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return QuitConfirm(v)
		//case "tab":
		case "shift+tab":
			return formTab1(v.inputs)
		default:
			// 타이핑·←→ 등 행 이동이 아닌 키는 포커스된 입력에 위임한다.
			return v, v.focus.Update(msg)
		}
	default:
		// 커서 깜빡임 등 키 외 메시지도 포커스된 입력에 위임한다.
		return v, v.focus.Update(msg)
	}
}
func (v *viewFormExampleTab2) View() tea.View {
	return tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		"[ summary ] [ tab1 ] [>tab2<]",
		//
		v.inputs.Habby.String(),
	))
}
