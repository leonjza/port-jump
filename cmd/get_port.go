package cmd

import (
	"errors"
	"fmt"
	"port-jump/internal/options"
	"port-jump/pkg/totp"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// portCmd represents the sshPort command
var portCmd = &cobra.Command{
	Use:   "port",
	Short: "Generate an SSH port to use.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := portCmdValidator(cmd); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetInt("port")
		var j *options.PortJump

		for _, jump := range opts.Jumps {
			if jump.DstPort == target {
				j = jump
			}
		}

		if j == nil {
			log.Error().Int("target", target).Msg("no configuration matching target found")
			return
		}

		totp, err := totp.NewTotp(j.SharedSecret, j.Interval)
		if err != nil {
			log.Error().Err(err).Msg("failed to get totp generator handle")
			return
		}

		port, err := totp.GenerateTCPPort()
		if err != nil {
			log.Error().Err(err).Msg("failed to generate TCP port")
			return
		}

		fmt.Printf("%d\n", port)
	},
}

func portCmdValidator(cmd *cobra.Command) error {
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return err
	}

	if port == 0 {
		return errors.New("port needs to be specified")
	}

	return nil
}

func init() {
	getCmd.AddCommand(portCmd)

	portCmd.Flags().IntP("port", "p", 0, "Destination port to use from configuration file")
}
