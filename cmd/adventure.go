package cmd

/*
Copyright © 2023 Kevin.Jayne@iCloud.com
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Character represents a player in the adventure
type Character struct {
	Name        string
	Description string
	HP          float64
	MP          float64
	Level       int64
	Strength    float64
	Defense     float64
	Dexterity   float64
	Intellect   float64
	Hunger      float64
}

var player Character

var adventureSystemMessage = `
You are the narrator and won't accept any answers that are not relevant to the current story or anything that wasn't mentioned yet.

Allow the user to do whatever they want keep it fantasy and fun.
Impart only minor rules to keep the story flowing.
Throw a few twists in the story but keep with the users original intent.

`

var generateImages = false
var adventureMessages []openai.ChatCompletionMessageParamUnion
var adventureStage int // 0 = name, 1 = description, 2 = playing

// adventureCmd represents the adventure command
var adventureCmd = &cobra.Command{
	Use:   "adventure",
	Short: "lets you dive into a captivating text adventure",
	Long:  `immerses you in a dynamic virtual story. Through text prompts, you'll make choices that lead your character through a series of challenges and decisions. Each choice you make affects the storyline's development, creating a unique and interactive narrative experience. Get ready to explore, solve puzzles, and shape the adventure's outcome entirely through your imagination and decisions.`,
	Run: func(cmd *cobra.Command, args []string) {
		adventureStage = 0
		player = Character{}
		adventureMessages = []openai.ChatCompletionMessageParamUnion{}

		p := tea.NewProgram(
			newChatHistoryModel(ChatHistoryConfig{
				Title:          "⚔️  Ponder Adventure ⚔️",
				Placeholder:    "Enter your name...",
				InitialMessage: "Welcome adventurer! Please type your name.",
				UserLabel:      "Unknown Adventurer: ",
				AssistantLabel: "Narrator:",
				UserColor:      "86",
				AssistantColor: "212",
				CustomHandler:  adventureHandler,
			}),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		if _, err := p.Run(); err != nil {
			catchErr(err, "fatal")
		}
	},
}

func init() {
	rootCmd.AddCommand(adventureCmd)
	adventureCmd.Flags().BoolVarP(&generateImages, "images", "i", false, "Generate Images")
}

func adventureHandler(m *chatHistoryModel, userInput string) tea.Cmd {
	switch adventureStage {
	case 0: // Name input
		player = Character{
			Name:        userInput,
			Description: "",
			HP:          100,
			MP:          100,
			Level:       1,
			Strength:    1,
			Defense:     1,
			Dexterity:   1,
			Intellect:   1,
			Hunger:      0,
		}
		adventureStage = 1
		m.textarea.Placeholder = "Describe your character..."

		// Update the user label to use the player's name for future messages
		m.config.UserLabel = "🗡️  " + player.Name + " 🛡️: "

		// Update the name entry message to avoid duplication (label will be prepended automatically)
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].role == "user" {
			m.messages[len(m.messages)-1].content = ""
		}

		// Add narrator response immediately (no API call needed)
		welcomeMsg := "Welcome " + player.Name + "! Now, please describe your character. Be as detailed as you like.\nYou can include their appearance, personality, background, skills, or anything else that defines them."
		return func() tea.Msg {
			return responseMsg{content: welcomeMsg, audio: nil, err: nil}
		}

	case 1: // Description input
		player.Description = userInput
		adventureStage = 2
		m.textarea.Placeholder = "What do you do?..."

		// Initialize adventure with character
		return func() tea.Msg {
			playerString, err := json.Marshal(player)
			if err != nil {
				return responseMsg{err: err}
			}

			adventureSystemMessage = adventureSystemMessage + "\n YOUR STARTING CHARACTER STATS:\n" + string(playerString)
			adventureMessages = []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(adventureSystemMessage),
			}

			response, audio := adventureResponse("My name is " + player.Name + " start adventure")
			return responseMsg{content: response, audio: audio, err: nil}
		}

	case 2: // Playing the adventure
		return func() tea.Msg {
			// Manage message history to prevent it from getting too long
			// Keep system message and last 20 messages
			if len(adventureMessages) > 21 {
				adventureMessages = append(adventureMessages[:1], adventureMessages[len(adventureMessages)-20:]...)
			}

			response, audio := adventureResponse(userInput)
			return responseMsg{content: response, audio: audio, err: nil}
		}
	}

	return nil
}

func adventureResponse(prompt string) (string, []byte) {
	var audio []byte
	spinner, _ = ponderSpinner.Start()
	response := adventureChat(prompt)
	if narrate {
		audio = tts(response)
	}
	spinner.Stop()
	if generateImages {
		go adventureImage(response)
	}
	return response, audio
}

func adventureChat(prompt string) string {
	adventureMessages = append(adventureMessages, openai.UserMessage(prompt))

	oaiResponse, err := ai.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: adventureMessages,
		Model:    viper.GetString("openAI_chat_model"),
	})
	catchErr(err)

	assistantMessage := oaiResponse.Choices[0].Message.Content
	adventureMessages = append(adventureMessages, openai.AssistantMessage(assistantMessage))
	return assistantMessage
}

func adventureImage(prompt string) {
	fmt.Println("🖼  Creating Image...")
	res, err := ai.Images.Generate(context.Background(), openai.ImageGenerateParams{
		Prompt: prompt,
		Model:  openai.ImageModel(viper.GetString("openAI_image_model")),
		Size:   openai.ImageGenerateParamsSize(viper.GetString("openAI_image_size")),
		N:      openai.Int(1),
	})
	if err != nil {
		fmt.Println("❌ Error generating image:", err)
		return
	}

	url := res.Data[0].URL

	promptFormatted := formatPrompt(prompt)
	filePath := viper.GetString("openAI_image_downloadPath")
	currentUser, err := user.Current()
	homeDir := currentUser.HomeDir
	catchErr(err)
	if filePath == `~` || strings.HasPrefix(filePath, "~") { // Replace ~ with home directory
		filePath = strings.Replace(filePath, "~", homeDir, 1)
	}

	fileName := promptFormatted[0:9] + strconv.Itoa(0) + ".jpg"
	fullFilePath := filepath.Join(filePath, fileName)
	// Create the directory (if it doesn't exist)
	err = os.MkdirAll(filePath, os.ModePerm)
	catchErr(err)
	// fmt.Printf("💾 Downloading Image:")
	url = httpDownloadFile(url, fullFilePath)
	// fmt.Printf(" \"%s\"\n", url)
	// fmt.Println("💻 Opening Image...")
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform for opening files: %s", runtime.GOOS)
	}
	if err != nil {
		trace()
		fmt.Println(err)
	}
}
