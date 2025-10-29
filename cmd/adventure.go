package cmd

/*
Copyright © 2023 Kevin.Jayne@iCloud.com
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vekjja/goai"
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
var adventureMessages = []goai.Message{}

// adventureCmd represents the adventure command
var adventureCmd = &cobra.Command{
	Use:   "adventure",
	Short: "lets you dive into a captivating text adventure",
	Long:  `immerses you in a dynamic virtual story. Through text prompts, you'll make choices that lead your character through a series of challenges and decisions. Each choice you make affects the storyline's development, creating a unique and interactive narrative experience. Get ready to explore, solve puzzles, and shape the adventure's outcome entirely through your imagination and decisions.`,
	Run: func(cmd *cobra.Command, args []string) {
		startAdventure()
	},
}

func init() {
	rootCmd.AddCommand(adventureCmd)
	adventureCmd.Flags().BoolVarP(&generateImages, "images", "i", false, "Generate Images")
}

func adventureChat(prompt string) string {
	adventureMessages = append(adventureMessages, goai.Message{
		Role:    "user",
		Content: prompt,
	})
	oaiResponse, err := ai.ChatCompletion(adventureMessages)
	catchErr(err)
	adventureMessages = append(adventureMessages, goai.Message{
		Role:    "assistant",
		Content: oaiResponse.Choices[0].Message.Content,
	})
	return oaiResponse.Choices[0].Message.Content
}

func adventureImage(prompt string) {
	fmt.Println("🖼  Creating Image...")
	res, err := ai.ImageGen(prompt, viper.GetString("openAI_image_model"), viper.GetString("openAI_image_size"), 1)
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

func narratorSay(text string) {
	fmt.Println()
	if narrate {
		audioData := tts(text)
		spinner.Stop()
		fmt.Println("🗣️  Narrator: ", text)
		if audioData != nil {
			playAudio(audioData)
		}
	} else {

		fmt.Println("🗣️  Narrator: ", text)
		spinner.Stop()
	}
}

func getPlayerInput(player *Character) string {
	fmt.Print("\n🗡️  " + player.Name + "🛡️ : ")
	playerInput, err := getUserInput("What do you do?")
	catchErr(err)
	return playerInput
}

func totalMessageCharacters() int {
	totalCharacters := 0
	for _, message := range adventureMessages {
		totalCharacters += len(message.Content)
	}
	return totalCharacters
}

func startAdventure() {

	spinner, _ = ponderSpinner.WithSequence(moonSequence...).Start()
	narratorSay("Please type your name.")
	fmt.Print("🗡️  Your Name: ")
	playerName, err := getUserInput("Enter your name...")
	catchErr(err)

	player = Character{
		Name:        playerName,
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

	// spinner, _ = ponderSpinner.WithSequence(moonSequence...).Start()
	narratorSay("Welcome " + player.Name + ", to the world of adventure! Describe your character, be as detailed as you like.")
	playerDescription := getPlayerInput(&player)
	player.Description = playerDescription

	spinner, _ = ponderSpinner.WithSequence(moonSequence...).Start()
	playerString, err := json.Marshal(player)
	catchErr(err)

	adventureSystemMessage = adventureSystemMessage + "\n YOUR STARTING CHARACTER STATS:\n" + string(playerString)

	adventureMessages = []goai.Message{{
		Role:    "system",
		Content: adventureSystemMessage,
	}}

	startMessage := adventureChat("My name is " + player.Name + " start adventure")
	narratorSay(startMessage)
	if generateImages {
		adventureImage(startMessage)
	}

	for {

		if totalMessageCharacters() > 4096 {
			adventureMessages = append(adventureMessages[:2], adventureMessages[3:]...)
		}

		playerInput := getPlayerInput(&player)
		spinner, _ = ponderSpinner.WithSequence(moonSequence...).Start()
		adventureResponse := adventureChat(playerInput)
		narratorSay(adventureResponse)
		if generateImages {
			adventureImage(adventureResponse)
		}
	}
}
