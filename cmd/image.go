package cmd

/*
Copyright © 2023 Kevin.Jayne@iCloud.com
*/

import (
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
)

var open, download bool
var n int = 1

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Generate an image from a prompt",
	Long:  ``,
	// Args: func(cmd *cobra.Command, args []string) error {
	// 	return checkArgs(args)
	// },
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
	imageCmd.Flags().IntVarP(&n, "n", "n", 1, "Number of images to generate")
}

func createImage(prompt string) {
	fmt.Println("🖼  Creating Image...")
	res, err := ai.ImageGen(prompt, viper.GetString("openAI_image_model"), viper.GetString("openAI_image_size"), n)

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
