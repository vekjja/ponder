package cmd

/*
Copyright © 2023 Kevin.Jayne@iCloud.com
*/

import (
	tea "github.com/charmbracelet/bubbletea"
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
	Run: func(cmd *cobra.Command, args []string) {
		var text string
		if len(args) > 0 {
			text = args[0]
			prompt = text
		}
		p := tea.NewProgram(
			initialChatHistoryModel(),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		if _, err := p.Run(); err != nil {
			catchErr(err, "fatal")
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
