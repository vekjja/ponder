package cmd

/*
Copyright © 2023 Kevin.Jayne@iCloud.com
*/

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/vekjja/goai"
)

func init() {
	rootCmd.AddCommand(chatCmd)
}

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Open ended chat with OpenAI",
	Long:  ``,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkArgs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if convo {
			// Use the interactive TUI for conversation mode
			p := tea.NewProgram(
				initialChatHistoryModel(),
				tea.WithAltScreen(),
			)
			if _, err := p.Run(); err != nil {
				catchErr(err, "fatal")
			}
		} else {
			// Single response mode (no TUI)
			response, audio := chatResponse(prompt)
			syntaxHighlight(response)
			if narrate {
				playAudio(audio)
			}
		}
	},
}

func chatResponse(prompt string) (string, []byte) {
	var audio []byte
	var response string
	spinner, _ = ponderSpinner.Start()
	response = chatCompletion(prompt)
	if narrate {
		audio = tts(response)
	}
	spinner.Stop()
	return response, audio
}

func chatCompletion(prompt string) string {
	ponderMessages = append(ponderMessages, goai.Message{
		Role:    "user",
		Content: prompt,
	})

	// Send the messages to OpenAI
	res, err := ai.ChatCompletion(ponderMessages)
	catchErr(err, "fatal")
	ponderMessages = append(ponderMessages, goai.Message{
		Role:    "assistant",
		Content: res.Choices[0].Message.Content,
	})
	return res.Choices[0].Message.Content
}

// Styles for the chat UI
var (
	userStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	assistantStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	systemStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
)

// chatMessage represents a single message in the conversation
type chatMessage struct {
	role    string // "user", "assistant", "system"
	content string
}

// responseMsg is sent when the API returns a response
type responseMsg struct {
	content string
	audio   []byte
	err     error
}

// chatHistoryModel is the main model for the chat interface
type chatHistoryModel struct {
	viewport viewport.Model
	textarea textarea.Model
	messages []chatMessage
	width    int
	height   int
	ready    bool
	waiting  bool // true when waiting for API response
	err      error
}

func initialChatHistoryModel() chatHistoryModel {
	ta := textarea.New()
	ta.Placeholder = "Type your message here..."
	ta.Focus()
	ta.CharLimit = 10000
	ta.ShowLineNumbers = false

	// Add initial prompt as first user message if provided
	messages := []chatMessage{}
	if prompt != "" {
		messages = append(messages, chatMessage{
			role:    "user",
			content: prompt,
		})
	}

	return chatHistoryModel{
		textarea: ta,
		messages: messages,
		waiting:  prompt != "", // If we have an initial prompt, start waiting
	}
}

func (m chatHistoryModel) Init() tea.Cmd {
	if m.waiting && len(m.messages) > 0 {
		// Send initial prompt
		return tea.Batch(textarea.Blink, m.getResponseCmd(m.messages[0].content))
	}
	return textarea.Blink
}

func (m chatHistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// Initialize viewport with proper dimensions
			headerHeight := 2
			footerHeight := 6 // Space for textarea + instructions
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true

			// Set textarea dimensions
			m.textarea.SetWidth(msg.Width - 4)
			m.textarea.SetHeight(3)
		} else {
			m.viewport.Width = msg.Width
			headerHeight := 2
			footerHeight := 6
			m.viewport.Height = msg.Height - headerHeight - footerHeight
			m.textarea.SetWidth(msg.Width - 4)
		}

		// Update viewport content
		m.viewport.SetContent(m.renderMessages())

	case responseMsg:
		m.waiting = false
		if msg.err != nil {
			m.err = msg.err
			m.messages = append(m.messages, chatMessage{
				role:    "system",
				content: fmt.Sprintf("Error: %v", msg.err),
			})
		} else {
			m.messages = append(m.messages, chatMessage{
				role:    "assistant",
				content: msg.content,
			})
		}
		// Update viewport first to show the text
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

		// Play audio after text is displayed
		if msg.err == nil && narrate && msg.audio != nil {
			go playAudio(msg.audio)
		}
		return m, nil

	case tea.KeyMsg:
		// Don't handle keys while waiting for response
		if m.waiting {
			if msg.Type == tea.KeyCtrlC {
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlD:
			// Submit the message
			if strings.TrimSpace(m.textarea.Value()) != "" {
				userMsg := strings.TrimSpace(m.textarea.Value())
				m.messages = append(m.messages, chatMessage{
					role:    "user",
					content: userMsg,
				})
				m.textarea.Reset()
				m.waiting = true
				m.viewport.SetContent(m.renderMessages())
				m.viewport.GotoBottom()
				return m, m.getResponseCmd(userMsg)
			}
			return m, nil
		}

	case error:
		m.err = msg
		return m, nil
	}

	// Update textarea or viewport based on focus
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

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Render("💭 Ponder Chat") + "\n\n"

	instructions := systemStyle.Render("Ctrl+D to send | Ctrl+C to quit")
	if m.waiting {
		instructions = systemStyle.Render("⏳ Waiting for response... | Ctrl+C to quit")
	}

	return header + m.viewport.View() + "\n\n" + m.textarea.View() + "\n" + instructions
}

// renderMessages converts the message history to a formatted string
func (m chatHistoryModel) renderMessages() string {
	var b strings.Builder

	for _, msg := range m.messages {
		switch msg.role {
		case "user":
			b.WriteString(userStyle.Render("You: "))
			b.WriteString(msg.content)
		case "assistant":
			b.WriteString(assistantStyle.Render("Ponder: ") + "\n")
			// Apply syntax highlighting to assistant messages
			highlighted := syntaxHighlightString(msg.content)
			b.WriteString(highlighted)
		case "system":
			b.WriteString(systemStyle.Render(msg.content))
		}
		b.WriteString("\n\n")
	}

	return b.String()
}

// getResponseCmd calls the AI and returns the response as a Cmd
func (m chatHistoryModel) getResponseCmd(userPrompt string) tea.Cmd {
	return func() tea.Msg {
		response, audio := chatResponse(userPrompt)
		return responseMsg{
			content: response,
			audio:   audio,
			err:     nil,
		}
	}
}

// simpleInputModel is a simple editor for single inputs
type simpleInputModel struct {
	textarea textarea.Model
	err      error
}

func (m simpleInputModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m simpleInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlD:
			// Submit
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m simpleInputModel) View() string {
	return fmt.Sprintf(
		"%s\n",
		m.textarea.View(),
	)
}

func getUserInput(placeholder string) (string, error) {
	// Create and configure the textarea
	ti := textarea.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 10000
	ti.SetWidth(80)
	ti.SetHeight(3)
	ti.ShowLineNumbers = false

	// Create the model
	m := simpleInputModel{
		textarea: ti,
		err:      nil,
	}

	// Run the program
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		trace()
		return "", fmt.Errorf("error running editor: %w", err)
	}

	// Get the final text
	if fm, ok := finalModel.(simpleInputModel); ok {
		result := strings.TrimSpace(fm.textarea.Value())
		if verbose > 0 {
			trace()
			fmt.Println(result)
		}
		return result, nil
	}

	return "", fmt.Errorf("unexpected model type")
}
