package cmd

import (
	"fmt"
	"port-jump/internal/secrets"

	"github.com/spf13/cobra"
)

// secretCmd represents the secret command
var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Generate some secrets to use in jumps",
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := cmd.Flags().GetInt("count")
		for range count {
			secret, _ := secrets.GenerateTOTPSecret(16)
			fmt.Println(secret)
		}
	},
}

func init() {
	rootCmd.AddCommand(secretCmd)

	secretCmd.PersistentFlags().IntP("count", "c", 8, "Number of secrets to generate")
}
