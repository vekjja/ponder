package cmd

/*
Copyright © 2023 Kevin.Jayne@iCloud.com
*/

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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

		var err error
		if convo {
			for {
				response, _ := chatResponse(prompt)
				// response, audio := chatResponse(prompt)
				fmt.Println("\nPonder:")
				syntaxHighlight(response)
				// if narrate {
				// 	playAudio(audio)
				// }
				fmt.Print("\nYou:\n  ")
				prompt, err = getUserInput()
				catchErr(err, "warn")
			}
		} else {
			response, _ := chatResponse(prompt)
			// response, audio := chatResponse(prompt)
			syntaxHighlight(response)
			// if narrate {
			// 	playAudio(audio)
			// }
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

func getUserInput() (string, error) {
	// ReadString will block until the delimiter is entered
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		trace()
		return "", err
	}
	// remove the delimiter from the string
	input = strings.TrimSuffix(input, "\n")
	if verbose > 0 {
		trace()
		fmt.Println(input)
	}
	return input, nil
}
