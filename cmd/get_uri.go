package cmd

import (
	"errors"
	"fmt"
	"port-jump/internal/options"
	"port-jump/pkg/hotp"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// uriCmd represents the uri command
var uriCmd = &cobra.Command{
	Use:   "uri",
	Short: "Generate a URI to use",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := uriCmdValidator(cmd); err != nil {
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

		uri, _ := cmd.Flags().GetString("uri")
		url, _ := cmd.Flags().GetString("url")

		totp, err := hotp.NewTotp(j.SharedSecret, j.Interval)
		if err != nil {
			log.Error().Err(err).Msg("failed to get totp generator handle")
			return
		}

		port, err := totp.GenerateTCPPort()
		if err != nil {
			log.Error().Err(err).Msg("failed to generate TCP port")
			return
		}

		fmt.Printf("%s%s:%d/\n", uri, url, port)
	},
}

func uriCmdValidator(cmd *cobra.Command) error {
	uri, err := cmd.Flags().GetString("uri")
	if err != nil {
		return err
	}

	if uri == "" {
		return errors.New("--uri cannot be empty")
	}

	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	if url == "" {
		return errors.New("--url cannot be empty")
	}

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
	getCmd.AddCommand(uriCmd)

	uriCmd.PersistentFlags().StringP("uri", "", "https://", "The URI handler to use.")
	uriCmd.PersistentFlags().StringP("url", "", "", "The URL to use.")
	uriCmd.Flags().IntP("port", "p", 0, "Destination port to use from configuration file")
}
