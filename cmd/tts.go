package cmd

/*
Copyright © 2024 Kevin Jayne <kevin.jayne@icloud.com>
*/

import (
	"bytes"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var audioFile,
	voice string

// ttsCmd represents the tts command
var ttsCmd = &cobra.Command{
	Use:   "tts",
	Short: "OpenAI Text to Speech API - TTS",
	Long: `OpenAI Text to Speech API - TTS
	You can use the TTS API to generate audio from text.
	`,

	Run: func(cmd *cobra.Command, args []string) {
		var text string
		if len(args) > 0 {
			text = args[0]
			prompt = text
		}
		// Open the chat history model for interactive TTS
		p := tea.NewProgram(
			initialTTSHistoryModel(),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		if _, err := p.Run(); err != nil {
			catchErr(err, "fatal")
		}
	},
}

func init() {
	rootCmd.AddCommand(ttsCmd)
	ttsCmd.Flags().StringVarP(&audioFile, "file", "f", "", "File to save audio to")
}

func initialTTSHistoryModel() chatHistoryModel {
	return newChatHistoryModel(ChatHistoryConfig{
		Title:           "🔊 Text to Speech",
		Placeholder:     "Enter text to convert to speech...",
		UserLabel:       "Text: ",
		AssistantLabel:  "Playing Audio",
		UserColor:       userColor,
		AssistantColor:  assistantColor,
		ResponseHandler: ttsResponse,
	})
}

func ttsResponse(text string) (string, []byte) {
	spinner, _ = ponderSpinner.Start()
	audio := tts(text)
	spinner.Stop()
	if audio != nil {
		go playAudio(audio)
	}
	return "", nil
}

func tts(text string) []byte {
	ai.Voice = voice
	audioData, err := ai.TTS(text)
	catchErr(err, "fatal")
	if audioFile != "" {
		file, err := os.Create(audioFile)
		catchErr(err)
		defer file.Close()
		_, err = io.Copy(file, bytes.NewReader(audioData))
		catchErr(err)
		return nil
	}
	return audioData
}
