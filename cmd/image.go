/*
Copyright © 2023 NAME HERE Kevin.Jayne@iCloud.com
*/
package cmd

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
var imageFile string
var n int

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Generate an image from a prompt",
	Long:  ``,
	Args: func(cmd *cobra.Command, args []string) error {
		return checkArgs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		createImage(prompt, imageFile)

	},
}

func init() {
	rootCmd.AddCommand(imageCmd)
	imageCmd.Flags().BoolVarP(&download, "download", "d", false, "Download image(s) to local directory")
	imageCmd.Flags().BoolVarP(&open, "open", "o", false, "Open image in system default viewer")
	imageCmd.Flags().IntVarP(&n, "n", "n", 1, "Number of images to generate")
	imageCmd.Flags().StringVarP(&imageFile, "file", "f", "", "Image file to edit")
}

func createImage(prompt, imageFile string) {
	fmt.Println("🖼  Creating Image...")
	res := ai.ImageGen(prompt, imageFile, n)

	fmt.Println("URLs:" + res.Data[0].URL)

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
