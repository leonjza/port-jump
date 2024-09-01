package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Work with port-jump configurations",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
