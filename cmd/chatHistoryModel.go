package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type responseMsg struct {
	content string
	audio   []byte
	err     error
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
}

func initialChatHistoryModel() chatHistoryModel {
	ta := textarea.New()
	ta.Placeholder = "Enter your message here..."
	ta.Focus()
	ta.CharLimit = 10000
	ta.ShowLineNumbers = false

	m := chatHistoryModel{textarea: ta}

	if prompt != "" {
		m.messages = append(m.messages, struct{ role, content string }{"user", prompt})
		m.waiting = true
	}

	return m
}

func (m chatHistoryModel) Init() tea.Cmd {
	if m.waiting {
		return tea.Batch(textarea.Blink, func() tea.Msg {
			response, audio := chatResponse(m.messages[0].content)
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
		// Layout: title(1) + viewport(h) + textarea(3) + help(1) = msg.Height
		// So: h = msg.Height - 5
		textareaHeight := 3
		h := msg.Height - textareaHeight - 2 // -2 for title and help
		if h < 1 {
			h = 1
		}
		if !m.ready {
			m.viewport = viewport.New(msg.Width, h)
			m.textarea.SetWidth(msg.Width - 4)
			m.textarea.SetHeight(textareaHeight)
			m.ready = true
		} else {
			m.viewport.Width, m.viewport.Height = msg.Width, h
			m.textarea.SetWidth(msg.Width - 4)
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
				return m, func() tea.Msg {
					response, audio := chatResponse(userMsg)
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

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("💭 Ponder Chat")
	helpLine := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true).Render(help)

	return fmt.Sprintf("%s\n%s\n%s\n%s", title, m.viewport.View(), m.textarea.View(), helpLine)
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

	for i, msg := range m.messages {
		if i > 0 {
			b.WriteString("\n")
		}
		switch msg.role {
		case "user":
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render("You: "))
			b.WriteString(wrap.Render(msg.content))
		case "assistant":
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render("Ponder:") + "\n")
			b.WriteString(wrap.Render(syntaxHighlightString(msg.content)))
		case "system":
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true).Render(wrap.Render(msg.content)))
		}
		b.WriteString("\n")
	}
	return b.String()
}
