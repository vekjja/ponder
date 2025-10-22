package cmd

/*
Copyright © 2023 Kevin.Jayne@iCloud.com
*/

import (
	"github.com/spf13/cobra"
)

var removeCMDIds string

// discordCmd represents the discord command
var discordCmd = &cobra.Command{
	Use:   "discord-bot",
	Short: "Discord Chat Bot Integration",
	Long:  `Discord Chat Bot Integration utilizing Secure Gateway Websocket`,
	Run: func(cmd *cobra.Command, args []string) {
		initDiscord()
	},
}

func init() {
	rootCmd.AddCommand(discordCmd)
	discordCmd.Flags().StringVarP(&removeCMDIds, "deregister-commands", "D", "", "A comma separated list of command IDs to deregister")
}
