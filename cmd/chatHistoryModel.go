package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UI configuration constants
const (
	textareaHeight = 3
	textareaWidth  = 4 // padding for width
	titleLines     = 1
	helpLines      = 1
	charLimit      = 10000

	// Colors
	titleColor     = "212"
	userColor      = "86"
	assistantColor = "212"
	helpColor      = "240"
	systemColor    = "240"
)

type responseMsg struct {
	content string
	audio   []byte
	err     error
}

// ChatHistoryConfig allows customization of the chat history model
type ChatHistoryConfig struct {
	Title           string
	Placeholder     string
	InitialMessage  string
	UserLabel       string
	AssistantLabel  string
	UserColor       string
	AssistantColor  string
	ResponseHandler func(string) (string, []byte)
	CustomHandler   func(*chatHistoryModel, string) tea.Cmd // For multi-stage interactions
}

type chatHistoryModel struct {
	viewport viewport.Model
	textarea textarea.Model
	messages []struct {
		role    string
		content string
	}
	ready   bool
	waiting bool
	config  ChatHistoryConfig
}

func newChatHistoryModel(config ChatHistoryConfig) chatHistoryModel {
	ta := textarea.New()
	ta.Placeholder = config.Placeholder
	if ta.Placeholder == "" {
		ta.Placeholder = "Enter your message here..."
	}
	ta.Focus()
	ta.CharLimit = charLimit
	ta.ShowLineNumbers = false

	m := chatHistoryModel{
		textarea: ta,
		config:   config,
	}

	if config.InitialMessage != "" {
		m.messages = append(m.messages, struct{ role, content string }{"assistant", config.InitialMessage})
	}

	if prompt != "" {
		m.messages = append(m.messages, struct{ role, content string }{"user", prompt})
		m.waiting = true
	}

	return m
}

func initialChatHistoryModel() chatHistoryModel {
	return newChatHistoryModel(ChatHistoryConfig{
		Title:           "💭 Ponder Chat",
		Placeholder:     "Enter your message here...",
		UserLabel:       "You: ",
		AssistantLabel:  "Ponder:",
		UserColor:       userColor,
		AssistantColor:  assistantColor,
		ResponseHandler: chatResponse,
	})
}

func (m chatHistoryModel) Init() tea.Cmd {
	if m.waiting {
		return tea.Batch(textarea.Blink, func() tea.Msg {
			response, audio := m.config.ResponseHandler(m.messages[0].content)
			return responseMsg{content: response, audio: audio}
		})
	}
	return textarea.Blink
}

func (m chatHistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Layout: title + viewport + textarea + help = msg.Height
		// viewport height = msg.Height - (titleLines + textareaHeight + helpLines)
		h := msg.Height - (titleLines + textareaHeight + helpLines)
		if h < 1 {
			h = 1
		}
		if !m.ready {
			m.viewport = viewport.New(msg.Width, h)
			m.textarea.SetWidth(msg.Width - textareaWidth)
			m.textarea.SetHeight(textareaHeight)
			m.ready = true
		} else {
			m.viewport.Width, m.viewport.Height = msg.Width, h
			m.textarea.SetWidth(msg.Width - textareaWidth)
			m.textarea.SetHeight(textareaHeight)
		}
		m.viewport.SetContent(m.renderMessages())

	case responseMsg:
		m.waiting = false
		if msg.err != nil {
			m.messages = append(m.messages, struct{ role, content string }{"system", fmt.Sprintf("Error: %v", msg.err)})
		} else {
			m.messages = append(m.messages, struct{ role, content string }{"assistant", msg.content})
			if narrate && msg.audio != nil {
				go playAudio(msg.audio)
			}
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			stopAudio()
			return m, tea.Quit
		}
		if m.waiting {
			return m, nil
		}
		if msg.Type == tea.KeyCtrlD {
			if userMsg := strings.TrimSpace(m.textarea.Value()); userMsg != "" {
				m.messages = append(m.messages, struct{ role, content string }{"user", userMsg})
				m.textarea.Reset()
				m.waiting = true
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()

				// Use custom handler if provided, otherwise use default response handler
				if m.config.CustomHandler != nil {
					return m, m.config.CustomHandler(&m, userMsg)
				}

				return m, func() tea.Msg {
					response, audio := m.config.ResponseHandler(userMsg)
					return responseMsg{response, audio, nil}
				}
			}
		}
	}

	if !m.waiting {
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m chatHistoryModel) View() string {
	if !m.ready {
		return "\nInitializing..."
	}

	help := "↑/↓ scroll | Ctrl+D send | Ctrl+C quit"
	if m.waiting {
		help = "⏳ Waiting... | Ctrl+C quit"
	}

	title := m.config.Title
	if title == "" {
		title = "💭 Chat"
	}
	titleRendered := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(titleColor)).Render(title)
	helpLine := lipgloss.NewStyle().Foreground(lipgloss.Color(helpColor)).Italic(true).Render(help)

	return fmt.Sprintf("%s\n%s\n%s\n%s", titleRendered, m.viewport.View(), m.textarea.View(), helpLine)
}

func (m chatHistoryModel) renderMessages() string {
	if len(m.messages) == 0 {
		return "Start typing below..."
	}

	var b strings.Builder
	w := m.viewport.Width
	if w == 0 {
		w = 80
	}
	wrap := lipgloss.NewStyle().Width(w)

	userColorToUse := m.config.UserColor
	if userColorToUse == "" {
		userColorToUse = userColor
	}
	assistantColorToUse := m.config.AssistantColor
	if assistantColorToUse == "" {
		assistantColorToUse = assistantColor
	}

	for i, msg := range m.messages {
		if i > 0 {
			b.WriteString("\n")
		}
		switch msg.role {
		case "user":
			label := m.config.UserLabel
			if label == "" {
				label = "You: "
			}
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(userColorToUse)).Bold(true).Render(label))
			b.WriteString(wrap.Render(msg.content))
		case "assistant":
			label := m.config.AssistantLabel
			if label == "" {
				label = "Assistant:"
			}
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(assistantColorToUse)).Bold(true).Render(label) + "\n")
			b.WriteString(wrap.Render(syntaxHighlightString(msg.content)))
		case "system":
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(systemColor)).Italic(true).Render(wrap.Render(msg.content)))
		}
		b.WriteString("\n")
	}
	return b.String()
}
