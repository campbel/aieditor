package main

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/sashabaranov/go-openai"
)

var (
	client *openai.Client

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	codeStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			BorderTop(true).BorderBottom(true).BorderLeft(false).BorderRight(false)
)

func main() {
	// options parsig
	if len(os.Args) != 2 {
		log.Fatal("Please provide a file")
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
	p := tea.NewProgram(initialModel(os.Args[1], string(file)), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
	keys keyMap
	help help.Model

	textInput textinput.Model
	code      viewport.Model
	message   string
	file      *file

	height int
	width  int
}

type file struct {
	name    string
	content string
	display string
}

func newFile(name, content string) *file {
	f := &file{name: name}
	f.update(content)
	return f
}

func (f *file) update(content string) {
	f.content = content
	var b bytes.Buffer
	err := quick.Highlight(&b, content, filepath.Ext(f.name), "terminal16m", "solorized-dark")
	if err != nil {
		f.display = content
	}
	f.display = b.String()
}

func initialModel(name, content string) model {
	ti := textinput.New()
	ti.Placeholder = "Fix this code!"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 20

	code := viewport.New(0, 0)
	code.Style = codeStyle

	return model{
		textInput: ti,
		help:      help.New(),
		keys:      keys,
		code:      code,
		file:      newFile(name, content),
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			m.code.LineUp(1)
		case key.Matches(msg, m.keys.Down):
			m.code.LineDown(1)
		case key.Matches(msg, m.keys.Enter):
			input := m.textInput.Value()
			m.textInput.Reset()
			response, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
				Model:     "text-davinci-003",
				Prompt:    fmt.Sprintf("Modify the code below in the following way (don't include the code block in output): %s\n\n```%s\n%s\n```\n", input, getLanguage(m.file.name), m.file.content),
				MaxTokens: 2000,
			})
			if err != nil {
				m.message = err.Error()
			}
			m.file.update(response.Choices[0].Text)

		case key.Matches(msg, m.keys.Save):
			err := os.WriteFile(m.file.name, []byte(m.file.content), 0644)
			if err != nil {
				m.message = err.Error()
			} else {
				m.message = "Saved"
			}
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	// We handle errors just like any other message
	case errMsg:
		m.message = msg.Error()
		return m, nil
	}

	m.sizeInputs()
	m.code.SetContent(m.file.display)
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *model) sizeInputs() {
	m.code.Height = m.height - 6
	m.code.Width = m.width - 2
	m.textInput.Width = m.width - 2
}

func (m model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Left,
			titleStyle.Width(len(m.file.name)).Render(m.file.name),
			messageStyle.Width(m.width-len(m.file.name)).Render(m.message),
		),
		"",
		m.textInput.View(),
		m.code.View(),
		m.help.View(m.keys),
	)
}

func getLanguage(filename string) string {
	extension := filepath.Ext(filename)
	switch extension {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".java":
		return "java"
	case ".cpp":
		return "cpp"
	case ".c":
		return "c"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	}
	return "text"
}
