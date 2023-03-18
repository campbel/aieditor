package main

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/sashabaranov/go-openai"

	"github.com/campbel/aieditor/app"
	"github.com/campbel/aieditor/diff"
	"github.com/campbel/aieditor/log"
)

var (
	client *openai.Client

	borderColor = lipgloss.Color("63")

	codeStyle = lipgloss.NewStyle()

	linesStyle = lipgloss.NewStyle().
			Width(5).
			AlignHorizontal(lipgloss.Right).
			Foreground(lipgloss.Color("63")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			BorderRight(true).
			PaddingRight(2)

	program *tea.Program
)

func main() {
	// options parsig
	if len(os.Args) != 2 {
		log.Fatal("Please provide a file")
	}

	// initialize openai client
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	tokenBytes, err := os.ReadFile(filepath.Join(homedir, ".config", "openai", "token"))
	if err != nil {
		log.Fatal(err)
	}
	token := strings.TrimSpace(string(tokenBytes))
	client = openai.NewClient(token)

	// start the tea program
	program = tea.NewProgram(initialModel(os.Args[1]), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}

type state string

const (
	stateViewing   state = "viewing"
	stateInput     state = "input"
	stateLoading   state = "loading"
	stateComparing state = "comparing"
)

type model struct {
	state state

	keysDefault   app.KeyMap
	keysComparing app.CompareKeysMap
	help          help.Model

	spinner spinner.Model

	textInput textinput.Model

	code  viewport.Model
	lines viewport.Model

	message     string
	file        *app.FileBuffer
	input       string
	changes     []change
	changeIndex int

	height int
	width  int
}

type resultMsg struct {
	input       string
	suggestions []change
	err         error
}

type change struct {
	raw  string
	diff string
}

func newChange(raw, content string) change {
	d, _ := diff.Diff(content, raw)
	return change{
		raw:  raw,
		diff: d,
	}
}

type fileMsg struct{}

func initialModel(path string) model {
	input := textinput.New()
	input.Placeholder = "Fix this code!"
	input.CharLimit = 256
	input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	input.Prompt = "➜ "
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	code := viewport.New(0, 0)
	code.Style = codeStyle

	lines := viewport.New(0, 0)
	lines.Style = linesStyle

	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		state:         stateViewing,
		spinner:       s,
		textInput:     input,
		help:          help.New(),
		keysDefault:   app.Keys,
		keysComparing: app.CompareKeys,
		code:          code,
		lines:         lines,
		file:          app.NewFile(path),
	}
}

func (m model) Init() tea.Cmd {
	m.file.Watch(func() {
		program.Send(fileMsg{})
	})
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.updateSizes()
		m.updateContent()
	case fileMsg:
		m.updateContent()
	case resultMsg:
		if msg.err != nil {
			m.message = msg.err.Error()
		} else {
			m.input = msg.input
			m.changes = msg.suggestions
			m.state = stateComparing
			m.updateContent()
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		log.Debug("key pressed", "key", msg.String())
		switch m.state {
		case stateComparing:
			return m.updateCompareKeys(msg)
		default:
			return m.updateDefaultKeys(msg)
		}
	}
	if m.state == stateInput {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) updateCompareKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keysDefault.Up):
		m.code.LineUp(1)
		m.lines.LineUp(1)
	case key.Matches(msg, m.keysDefault.Down):
		m.code.LineDown(1)
		m.lines.LineDown(1)
	case key.Matches(msg, m.keysDefault.Top):
		m.code.GotoTop()
		m.lines.GotoTop()
	case key.Matches(msg, m.keysDefault.Bottom):
		m.code.GotoBottom()
		m.lines.GotoBottom()
	case key.Matches(msg, m.keysComparing.Next):
		m.changeIndex++
		if m.changeIndex >= len(m.changes) {
			m.changeIndex = 0
		}
		m.updateContent()
		return m, nil
	case key.Matches(msg, m.keysComparing.Prev):
		m.changeIndex--
		if m.changeIndex < 0 {
			m.changeIndex = len(m.changes) - 1
		}
		m.updateContent()
		return m, nil
	case key.Matches(msg, m.keysComparing.Accept):
		m.file.Set(m.changes[m.changeIndex].raw)
		m.state = stateViewing
		m.changes = nil
		m.input = ""
		m.changeIndex = 0
		m.updateContent()
		return m, nil
	case key.Matches(msg, m.keysComparing.Reject):
		// remove the element at m.changeIndex, and fix change index if its out of bounds
		m.changes = append(m.changes[:m.changeIndex], m.changes[m.changeIndex+1:]...)
		if m.changeIndex >= len(m.changes) {
			m.changeIndex = len(m.changes) - 1
		}
		if len(m.changes) == 0 {
			m.state = stateViewing
			m.changes = nil
			m.input = ""
			m.changeIndex = 0
		}
		m.updateContent()
		return m, nil
	case key.Matches(msg, m.keysComparing.Retry):
		m.state = stateLoading
		m.changes = nil
		m.input = ""
		m.changeIndex = 0
		m.updateContent()
		go m.fetchSuggestions(m.input)
	case key.Matches(msg, m.keysComparing.Exit):
		m.state = stateViewing
		m.changes = nil
		m.input = ""
		m.changeIndex = 0
		m.updateContent()
	}
	return m, nil
}

func (m *model) updateDefaultKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keysDefault.Focus):
		if m.state == stateViewing {
			m.textInput.Focus()
			m.state = stateInput
			return m, nil
		}
	case key.Matches(msg, m.keysDefault.Blur):
		if m.state == stateInput {
			m.textInput.Blur()
			m.state = stateViewing
			return m, nil
		}
	case key.Matches(msg, m.keysDefault.Up):
		m.code.LineUp(1)
		m.lines.LineUp(1)
	case key.Matches(msg, m.keysDefault.Down):
		m.code.LineDown(1)
		m.lines.LineDown(1)
	case key.Matches(msg, m.keysDefault.Top):
		m.code.GotoTop()
		m.lines.GotoTop()
	case key.Matches(msg, m.keysDefault.Bottom):
		m.code.GotoBottom()
		m.lines.GotoBottom()
	case key.Matches(msg, m.keysDefault.Enter):
		if m.state != stateLoading {
			m.state = stateLoading
			go m.fetchSuggestions(m.textInput.Value())
			m.textInput.Reset()
		}
		return m, nil
	case key.Matches(msg, m.keysDefault.Save):
		err := m.file.Save()
		if err != nil {
			m.message = err.Error()
		} else {
			m.message = "Saved"
		}
	case key.Matches(msg, m.keysDefault.Quit):
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *model) fetchSuggestions(input string) {
	log.Debug("sending request to openai", "input", input)
	response, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
		Model:     "text-davinci-003",
		Prompt:    fmt.Sprintf("Modify the code below in the following way (don't include the code block in output): %s\n\n```%s\n%s\n```\n", input, app.GetLanguage(m.file.Path()), m.file.Content()),
		MaxTokens: 2000,
		N:         3,
	})
	if err != nil {
		program.Send(resultMsg{err: err})
	} else {
		changes := []change{}
		for _, choice := range response.Choices {
			changes = append(changes, newChange(choice.Text, m.file.Content()))
		}
		program.Send(resultMsg{input: input, suggestions: changes})
	}
}

func (m *model) updateSizes() {
	log.Debug("updating sizes", "height", m.height, "width", m.width)
	m.code.Height = m.height - 5
	m.code.Width = m.width
	m.lines.Height = m.height - 3
	m.textInput.Width = m.width - 10 - len(m.file.Path())
}

func (m *model) updateContent() {
	log.Debug("updating content")
	content := m.file.Display()
	if m.state == stateComparing {
		if m.changes[m.changeIndex].diff != "" {
			content = m.changes[m.changeIndex].diff
		} else {
			content = app.Highlight(m.changes[m.changeIndex].raw, app.GetLanguage(m.file.Path()))
		}
	}
	m.code.SetContent(content)
	lines := ""
	for i := 0; i < len(strings.Split(content, "\n")); i++ {
		padding := ""
		if i < 9 {
			padding += " "
		}
		if i < 99 {
			padding += " "
		}
		if i > 0 {
			lines += "\n"
		}
		lines += fmt.Sprintf("%s%d", padding, i+1)
	}
	m.lines.SetContent(lines)
}

func (m model) View() string {
	headerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		BorderBottom(true)

	inputStyle := lipgloss.NewStyle().
		Width(m.width - len(m.file.Path()) - 1)

	filenameStyle := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Width(len(m.file.Path()) + 1).
		Bold(true)

	footerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		BorderTop(true)

	notificationStyle := lipgloss.NewStyle().
		Width(m.width).
		Foreground(lipgloss.Color("#0000ff"))

	if m.state == stateComparing {
		return lipgloss.JoinVertical(lipgloss.Top,
			headerStyle.Render(
				lipgloss.JoinHorizontal(lipgloss.Left,
					lipgloss.JoinHorizontal(lipgloss.Left,
						inputStyle.Foreground(lipgloss.Color("205")).Render(
							fmt.Sprintf("AI suggested change %d/%d", m.changeIndex+1, len(m.changes)),
						),
						filenameStyle.Render(m.file.Path()),
					),
				)),
			m.code.View(),
			footerStyle.Width(m.width).Render(
				lipgloss.JoinVertical(lipgloss.Top,
					m.helpView(),
					notificationStyle.Render(m.message),
				),
			),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		headerStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				lipgloss.JoinHorizontal(lipgloss.Left,
					inputStyle.Render(m.inputView()),
					filenameStyle.Render(m.file.Path()),
				),
			)),
		m.code.View(),
		footerStyle.Width(m.width).Render(
			lipgloss.JoinVertical(lipgloss.Top,
				m.helpView(),
				notificationStyle.Render(m.message),
			),
		),
	)
}

func (m *model) inputView() string {
	if m.state == stateLoading {
		return m.spinner.View() + " fetching code completions..."
	}
	if m.state == stateInput {
		return m.textInput.View()
	}
	return ""
}

func (m *model) helpView() string {
	switch m.state {
	case stateComparing:
		return m.help.View(m.keysComparing)
	default:
		return m.help.View(m.keysDefault)
	}
}
