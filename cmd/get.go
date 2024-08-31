package cmd

import (
	"github.com/spf13/cobra"
)

// getCmd represents the generate command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get port output to use in other programs",
}

func init() {
	rootCmd.AddCommand(getCmd)
}
