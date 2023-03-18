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
	"github.com/campbel/aieditor/log"
)

var (
	client *openai.Client

	borderColor = lipgloss.Color("63")

	codeStyle = lipgloss.NewStyle().
			PaddingLeft(1)

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

type model struct {
	keys app.KeyMap
	help help.Model

	spinner   spinner.Model
	loading   bool
	textInput textinput.Model

	code  viewport.Model
	lines viewport.Model

	message string
	file    *app.FileBuffer

	height int
	width  int
}

type resultMsg struct {
	content string
	err     error
}

type fileMsg struct{}

func initialModel(path string) model {
	input := textinput.New()
	input.Placeholder = "Fix this code!"
	input.CharLimit = 256
	input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	input.Prompt = "➜ "
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	input.Focus()

	code := viewport.New(0, 0)
	code.Style = codeStyle

	lines := viewport.New(0, 0)
	lines.Style = linesStyle

	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner:   s,
		textInput: input,
		help:      help.New(),
		keys:      app.Keys,
		code:      code,
		lines:     lines,
		file:      app.NewFile(path),
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
			m.file.Set(msg.content)
			m.updateContent()
		}
		m.loading = false
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		log.Debug("key pressed", "key", msg.String())
		switch {
		case key.Matches(msg, m.keys.Up):
			m.code.LineUp(1)
			m.lines.LineUp(1)
		case key.Matches(msg, m.keys.Down):
			m.code.LineDown(1)
			m.lines.LineDown(1)
		case key.Matches(msg, m.keys.Top):
			m.code.GotoTop()
			m.lines.GotoTop()
		case key.Matches(msg, m.keys.Bottom):
			m.code.GotoBottom()
			m.lines.GotoBottom()
		case key.Matches(msg, m.keys.Enter):
			if !m.loading {
				m.loading = true
				go func(input string) {
					log.Debug("sending request to openai", "input", input)
					response, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
						Model:     "text-davinci-003",
						Prompt:    fmt.Sprintf("Modify the code below in the following way (don't include the code block in output): %s\n\n```%s\n%s\n```\n", input, app.GetLanguage(m.file.Path()), m.file.Content()),
						MaxTokens: 2000,
					})
					if err != nil {
						program.Send(resultMsg{err: err})
					} else {
						program.Send(resultMsg{content: response.Choices[0].Text})
					}
				}(m.textInput.Value())
				m.textInput.Reset()
			}
			return m, nil
		case key.Matches(msg, m.keys.Save):
			err := m.file.Save()
			if err != nil {
				m.message = err.Error()
			} else {
				m.message = "Saved"
			}
		// case key.Matches(msg, m.keys.Undo):
		// 	m.file.Undo()
		// 	m.updateContent()
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	if !m.loading {
		var cmd tea.Cmd
		log.Debug("updating text input")
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) updateSizes() {
	log.Debug("updating sizes", "height", m.height, "width", m.width)
	m.code.Height = m.height - 5
	m.code.Width = m.width
	m.lines.Height = m.height - 3
	m.textInput.Width = m.width - 10 - len(m.file.Path())
}

func (m *model) updateContent() {
	log.Debug("updating content", "content", len(m.file.Display()))
	m.code.SetContent(m.file.Display())
	lines := ""
	for i := 0; i < len(strings.Split(m.file.Display(), "\n")); i++ {
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

	spacerStyle := lipgloss.NewStyle().
		Width(5).
		Height(1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		BorderRight(true)

	inputStyle := lipgloss.NewStyle().
		MarginLeft(1).
		Width(m.width - len(m.file.Path()) - 7)

	filenameStyle := lipgloss.NewStyle().
		Width(len(m.file.Path()) + 1).
		Bold(true)

	footerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		BorderTop(true)

	notificationStyle := lipgloss.NewStyle().
		Width(m.width).
		Foreground(lipgloss.Color("#0000ff"))

	return lipgloss.JoinVertical(lipgloss.Top,
		headerStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				spacerStyle.Render(""),
				lipgloss.JoinHorizontal(lipgloss.Left,
					inputStyle.Render(m.inputView()),
					filenameStyle.Render(m.file.Path()),
				),
			)),
		lipgloss.JoinHorizontal(lipgloss.Left,
			m.lines.View(),
			m.code.View(),
		),
		footerStyle.Width(m.width).Render(
			lipgloss.JoinVertical(lipgloss.Top,
				m.help.View(m.keys),
				notificationStyle.Render(m.message),
			),
		),
	)
}

func (m *model) inputView() string {
	if m.loading {
		return m.spinner.View() + " fetching code completions..."
	}
	return m.textInput.View()
}
