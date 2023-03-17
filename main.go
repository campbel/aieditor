package main

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	openai "github.com/sashabaranov/go-openai"

	"github.com/campbel/aieditor/app"
)

var (
	environment = os.Getenv("ENVIRONMENT")

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
	logger  log.Logger
)

func main() {
	// options parsig
	if len(os.Args) != 2 {
		log.Fatal("Please provide a file")
	}

	if environment == "development" {
		var (
			logFile *os.File
		)
		filename := "log.txt"
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			logFile, _ = os.Create(filename)
		} else {
			logFile, _ = os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0660)
		}
		logger = log.New(log.WithOutput(logFile), log.WithLevel(log.DebugLevel))
	} else {
		logger = log.New(log.WithLevel(log.InfoLevel))
	}

	// read the file
	file, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
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
	program = tea.NewProgram(initialModel(os.Args[1], string(file)), tea.WithAltScreen())
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
	file    *file

	height int
	width  int
}

type resultMsg struct {
	content string
	err     error
}

type file struct {
	name    string
	content []string
	display string
}

func newFile(name, content string) *file {
	f := &file{name: name}
	f.push(content)
	return f
}

func (f *file) undo() {
	if len(f.content) > 1 {
		f.content = f.content[1:]
	}
	f.update()
}

func (f *file) push(content string) {
	content = strings.TrimSpace(strings.Replace(content, "\t", "    ", -1))
	f.content = append([]string{content}, f.content...)
	f.update()
}

func (f *file) update() {
	var b bytes.Buffer
	err := quick.Highlight(&b, f.content[0], app.GetLanguage(f.name), "terminal16m", "dracula")
	if err != nil {
		f.display = f.content[0]
	}
	f.display = b.String()
}

func initialModel(name, content string) model {
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
		file:      newFile(name, content),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.updateSizes()
		m.updateContent()
	case resultMsg:
		if msg.err != nil {
			m.message = msg.err.Error()
		} else {
			m.file.push(msg.content)
			m.updateContent()
		}
		m.loading = false
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		logger.Debug("key pressed", "key", msg.String())
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
					logger.Debug("sending request to openai", "input", input)
					response, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
						Model:     "text-davinci-003",
						Prompt:    fmt.Sprintf("Modify the code below in the following way (don't include the code block in output): %s\n\n```%s\n%s\n```\n", input, app.GetLanguage(m.file.name), m.file.content),
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
			content := strings.TrimSpace(m.file.content[0] + "\n")
			err := os.WriteFile(m.file.name, []byte(content), 0644)
			if err != nil {
				m.message = err.Error()
			} else {
				m.message = "Saved"
			}
		case key.Matches(msg, m.keys.Undo):
			m.file.undo()
			m.updateContent()
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	if !m.loading {
		var cmd tea.Cmd
		logger.Debug("updating text input")
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) updateSizes() {
	logger.Debug("updating sizes", "height", m.height, "width", m.width)
	m.code.Height = m.height - 5
	m.code.Width = m.width
	m.lines.Height = m.height - 3
	m.textInput.Width = m.width - 10 - len(m.file.name)
}

func (m *model) updateContent() {
	logger.Debug("updating content", "content", len(m.file.display))
	m.code.SetContent(m.file.display)
	lines := ""
	for i := 0; i < len(strings.Split(m.file.display, "\n")); i++ {
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
		Width(m.width - len(m.file.name) - 7)

	filenameStyle := lipgloss.NewStyle().
		Width(len(m.file.name) + 1).
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
					filenameStyle.Render(m.file.name),
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
