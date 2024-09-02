package cmd

import (
	"os"
	"os/signal"
	"port-jump/internal/options"
	"port-jump/pkg/firewall"
	"port-jump/pkg/hotp"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// jumpCmd represents the jump command
var jumpCmd = &cobra.Command{
	Use:   "jump",
	Short: "Ensures a port jump remains configured.",
	Run: func(cmd *cobra.Command, args []string) {
		skip, err := cmd.Flags().GetBool("skip-cleanup")
		if err != nil {
			log.Error().Err(err).Msg("failed to parse skip-cleanup configuration")
			return
		}

		defer func() {
			if skip {
				return
			}

			if err := firewall.DeleteRules(); err != nil {
				log.Error().Err(err).Msg("failed to cleanup firewall rules")
			}
		}()

		// Channel to listen for termination signal
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

		if !haveJumps() {
			log.Error().Msg("there are no enabled jumps in the configuration file")
			skip = true // dont do any cleanups, we didnt create anything.
			return
		}

		// loop the configured jumps
		for _, jump := range opts.Jumps {
			if !jump.Enabled {
				continue
			}

			go func(j *options.PortJump) {
				jmpLog := log.With().Int("dst", j.DstPort).Bool("enabled", j.Enabled).Logger()

				portGen, err := hotp.NewTotp(j.SharedSecret, j.Interval)
				if err != nil {
					jmpLog.Error().Err(err).Msg("failed to get port generator for jump")
					return
				}

				var port = 0

				for {
					newPort, err := portGen.GenerateTCPPort()
					if err != nil {
						jmpLog.Error().Err(err).Msg("failed to get a tcp port from portGen")
						return
					}

					if newPort == port {
						time.Sleep(time.Millisecond * 500)
						continue
					}

					port = newPort
					if err := firewall.AddOrUpdateRedirect(port, j.DstPort); err != nil {
						jmpLog.Error().Err(err).Msg("failed to update nftables")
					}

					jmpLog.Info().Int("new-port", port).Msg("port jumped")
				}

			}(jump)
		}

		// block until we need to leave
		<-stopChan

		log.Info().Msg("exiting")
	},
}

// haveJumps checks if there are any enabled jumps
func haveJumps() bool {
	var enabled bool
	for _, jump := range opts.Jumps {
		if jump.Enabled {
			enabled = true
		}
	}

	return enabled
}

func init() {
	rootCmd.AddCommand(jumpCmd)

	jumpCmd.PersistentFlags().BoolP("skip-cleanup", "", false, "Do not cleanup the jump-port firewall table on exit")
}
