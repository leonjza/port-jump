package cmd

import (
	"os"

	"port-jump/internal/log"
	"port-jump/internal/options"
	"port-jump/internal/secrets"

	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var opts = options.NewOptions()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "port-jump",
	Short: "A proof-of-concept 'port-jump' tool. Change the port something listens on, over time.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := rootCmdValidator(cmd); err != nil {
			return err
		}

		if err := opts.Load(); err != nil {
			return err
		}

		log.Setup(opts.LogDebug)

		if len(opts.Jumps) == 0 {
			zlog.Warn().Msg("no configurations found. generating a disabled ssh example for you. check out the config file for details")
			s, err := secrets.GenerateTOTPSecret(16)
			if err != nil {
				return err
			}
			jmp, err := options.NewPortJump(22, s, int64(30), false)
			if err != nil {
				return err
			}

			opts.Jumps = append(opts.Jumps, jmp)
			opts.Save()
		}

		return nil
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func rootCmdValidator(cmd *cobra.Command) error {
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return err
	}

	opts.LogDebug = debug

	return nil
}

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "D", false, "debug")
}
