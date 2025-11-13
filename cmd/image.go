package cmd

/*
Copyright © 2023 Kevin.Jayne@iCloud.com
*/

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var open, download bool
var n int = 1

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Generate an image from a prompt",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var prompt string
		if len(args) > 0 {
			prompt = args[0]
		}
		createImage(prompt)
	},
}

func init() {
	rootCmd.AddCommand(imageCmd)
	imageCmd.Flags().BoolVarP(&download, "download", "d", false, "Download image(s) to local directory")
	imageCmd.Flags().BoolVarP(&open, "open", "o", false, "Open image in system default viewer")
	imageCmd.Flags().IntVarP(&n, "count", "c", 1, "Number of images to generate")
}

func createImage(prompt string) {
	fmt.Println("🖼  Creating Image...")
	res, err := ai.Images.Generate(context.Background(), openai.ImageGenerateParams{
		Prompt: prompt,
		Model:  openai.ImageModel(viper.GetString("openAI_image_model")),
		Size:   openai.ImageGenerateParamsSize(viper.GetString("openAI_image_size")),
		N:      openai.Int(int64(n)),
	})

	if err != nil {
		fmt.Println("❌ Error generating image:", err)
		return
	}

	for imgNum, data := range res.Data {
		url := data.URL
		fmt.Println("🌐 Image URL: " + url)

		if download { // Download image to local directory if download flag is set
			promptFormatted := formatPrompt(prompt)
			filePath := viper.GetString("openAI_image_downloadPath")
			currentUser, err := user.Current()
			homeDir := currentUser.HomeDir
			catchErr(err)
			if filePath == `~` || strings.HasPrefix(filePath, "~") { // Replace ~ with home directory
				filePath = strings.Replace(filePath, "~", homeDir, 1)
			}

			fileName := promptFormatted + strconv.Itoa(imgNum) + ".jpg"
			fullFilePath := filepath.Join(filePath, fileName)
			// Create the directory (if it doesn't exist)
			err = os.MkdirAll(filePath, os.ModePerm)
			catchErr(err)
			fmt.Printf("💾 Downloading Image:")
			url = httpDownloadFile(url, fullFilePath)
			fmt.Printf(" \"%s\"\n", url)
		}
		err := error(nil)
		if open { // Open image in browser if open flag is set
			fmt.Println("💻 Opening Image...")
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
	}
}
