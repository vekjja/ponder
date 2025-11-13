package cmd

/*
Copyright © 2023 Kevin Jayne <kevin.jayne@icloud.com>
*/

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vekjja/goai"
)

var ponderMessages = []goai.Message{}
var appVersion = "v0.4.3"
var ai *goai.Client

var verbose int

var convo,
	narrate bool

var prompt,
	configFile,
	openaiAPIKey,
	discordAPIKey string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ponder",
	Short: "Ponder OpenAI Chat Bot " + appVersion,
	Long: `
	Ponder
	GitHub: https://github.com/seemywingz/ponder
	App Version: ` + appVersion + `

  Ponder uses OpenAI's API to generate text responses to user input.
  Or whatever else you can think of. 🤔
	`,
	// Args: func(cmd *cobra.Command, args []string) error {
	// 	return checkArgs(args)
	// },
	Run: func(cmd *cobra.Command, args []string) {
		var prompt string
		if len(args) > 0 {
			prompt = args[0] // Use the first positional argument as the prompt
		}
		chatCmd.Run(cmd, []string{prompt})
	},
}

// func checkArgs(args []string) error {
// 	if convo && len(args) == 0 {
// 		// When --convo is used, no args are required
// 		return nil
// 	}
// 	// Otherwise, exactly one arg must be provided
// 	if len(args) != 1 {
// 		return fmt.Errorf("Prompt Required")
// 	}
// 	prompt = args[0]
// 	return nil
// }

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		CleanupAndExit()
		os.Exit(1)
	}
}

// CleanupAndExit performs cleanup before exiting the application
func CleanupAndExit() {
	stopAudio()
}

func init() {

	cobra.OnInitialize(viperConfig)

	rootCmd.MarkFlagRequired("prompt")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "verbose output (use -v, -vv, -vvv for more)")
	rootCmd.PersistentFlags().BoolVarP(&narrate, "narrate", "n", false, "Narrate the response using TTS and the default audio output")
	rootCmd.PersistentFlags().StringVar(&voice, "voice", "onyx", "Voice to use: alloy, ash, coral, echo, fable, onyx, nova, sage and shimmer")

	// Check for Required Environment Variables
	openaiAPIKey = os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" && verbose > 0 {
		fmt.Println("⚠️ OPENAI_API_KEY environment variable is not set, continuing without OpenAI API Key")
	}

	discordAPIKey = os.Getenv("DISCORD_API_KEY")
	if discordAPIKey == "" && verbose > 0 {
		fmt.Println("⚠️ DISCORD_API_KEY environment variable is not set, continuing without Discord API Key")
	}

}

func viperConfig() {
	// use spf13/viper to read config file

	viper.SetDefault("openAI_endpoint", "https://api.openai.com/v1/")

	viper.SetDefault("openAI_image_model", "dall-e-3")
	viper.SetDefault("openAI_image_size", "1024x1024")
	viper.SetDefault("openAI_image_downloadPath", "~/Ponder/Images/")

	viper.SetDefault("openAI_tts_model", "tts-1")
	viper.SetDefault("openAI_tts_voice", "onyx")
	viper.SetDefault("openAI_tts_speed", "1")
	viper.SetDefault("openAI_tts_responseFormat", "mp3")

	viper.SetDefault("openAI_voice", "onyx")
	viper.SetDefault("openAI_speed", "1")
	viper.SetDefault("openAI_responseFormat", "mp3")

	viper.SetDefault("openAI_chat_model", "gpt-4")
	viper.SetDefault("openAI_chat_systemMessage", "You are a helpful assistant.")

	viper.SetDefault("openAI_topP", "0.9")
	viper.SetDefault("openAI_frequencyPenalty", "0.0")
	viper.SetDefault("openAI_presencePenalty", "0.6")
	viper.SetDefault("openAI_temperature", "0")
	viper.SetDefault("openAI_maxTokens", "4096")

	viper.SetDefault("radio_notificationSound", "~/.ponder/audio/notify.mp3")

	viper.SetConfigName("config")        // name of config file (without extension)
	viper.SetConfigType("yaml")          // REQUIRED the config file does not have an extension
	viper.AddConfigPath(".")             // look for config in the working directory
	viper.AddConfigPath("./files")       // look for config in the working directory /files
	viper.AddConfigPath("$HOME/.ponder") // call multiple times to add many search paths

	if configFile != "" {
		viper.SetConfigFile(configFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("⚠️  Error Opening Config File:", err.Error(), "- Using Defaults")
	} else {
		if verbose > 0 {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

	ponderMessages = []goai.Message{{
		Role:    "developer",
		Content: viper.GetString("openAI_chat_systemMessage"),
	}}

	ai = &goai.Client{
		Endpoint:         viper.GetString("openAI_endpoint"),
		API_KEY:          openaiAPIKey,
		Verbose:          verbose,
		ImageSize:        viper.GetString("openAI_image_size"),
		User:             "ponder" + goai.HashAPIKey(openaiAPIKey),
		TopP:             viper.GetFloat64("openAI_topP"),
		ChatModel:        viper.GetString("openAI_chat_model"),
		ImageModel:       viper.GetString("openAI_image_model"),
		TTSModel:         viper.GetString("openAI_tts_model"),
		Voice:            viper.GetString("openAI_tts_voice"),
		Speed:            viper.GetFloat64("openAI_tts_speed"),
		ResponseFormat:   viper.GetString("openAI_tts_responseFormat"),
		MaxTokens:        viper.GetInt("openAI_maxTokens"),
		Temperature:      viper.GetFloat64("openAI_temperature"),
		FrequencyPenalty: viper.GetFloat64("openAI_frequencyPenalty"),
		PresencePenalty:  viper.GetFloat64("openAI_presencePenalty"),
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}
