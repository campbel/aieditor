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
	"github.com/charmbracelet/bubbles/spinner"
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

	program *tea.Program
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
	program = tea.NewProgram(initialModel(os.Args[1], string(file)), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
	keys keyMap
	help help.Model

	spinner   spinner.Model
	loading   bool
	textInput textinput.Model
	code      viewport.Model
	message   string
	file      *file

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
	content = strings.TrimSpace(strings.Replace(content, "\t", "    ", -1)) + "\n\n\n"
	f.content = append([]string{content}, f.content...)
	f.update()
}

func (f *file) update() {
	var b bytes.Buffer
	err := quick.Highlight(&b, f.content[0], getLanguage(f.name), "terminal16m", "dracula")
	if err != nil {
		f.display = f.content[0]
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

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner:   s,
		textInput: ti,
		help:      help.New(),
		keys:      keys,
		code:      code,
		file:      newFile(name, content),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case resultMsg:
		if msg.err != nil {
			m.message = msg.err.Error()
		} else {
			m.file.push(msg.content)
		}
		m.message = fmt.Sprintf("results are in! %d", len(msg.content))
		m.loading = false
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			m.code.LineUp(1)
		case key.Matches(msg, m.keys.Down):
			m.code.LineDown(1)
		case key.Matches(msg, m.keys.Top):
			m.code.GotoTop()
		case key.Matches(msg, m.keys.Bottom):
			m.code.GotoBottom()
		case key.Matches(msg, m.keys.Enter):
			if !m.loading {
				m.loading = true
				go func(input string) {
					response, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
						Model:     "text-davinci-003",
						Prompt:    fmt.Sprintf("Modify the code below in the following way (don't include the code block in output): %s\n\n```%s\n%s\n```\n", input, getLanguage(m.file.name), m.file.content),
						MaxTokens: 2000,
					})
					if err != nil {
						program.Send(resultMsg{err: err})
					} else {
						program.Send(resultMsg{content: response.Choices[0].Text})
					}
				}(m.textInput.Value())
				m.textInput.Reset()
				return m, nil
			}

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
			"\t",
			m.statusView(),
		),
		m.textInput.View(),
		m.code.View(),
		m.help.View(m.keys),
	)
}

func (m model) statusView() string {
	if m.loading {
		return m.spinner.View() + messageStyle.Width(m.width-len(m.file.name)).Render("loading...")
	}
	return messageStyle.Width(m.width - len(m.file.name)).Render(m.message)
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
