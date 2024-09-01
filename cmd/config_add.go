package cmd

import (
	"errors"
	"fmt"
	"port-jump/internal/options"
	"port-jump/internal/secrets"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new jump",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			portString     string
			intervalString string
			secret         string
			confirm        bool

			// converted values
			port     int
			interval int64
		)

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Destination port").
					Description("The local destination port where jumps should redirect to.").
					Placeholder("A port number.").
					Value(&portString).
					Validate(func(s string) error {
						p, err := strconv.Atoi(s)
						if err != nil {
							return errors.New("not a valid number")
						}

						if p < 1 || p > 65535 {
							return errors.New("please select a port between 1 and 65535")
						}

						if checkIfJumpExists(p) {
							return errors.New("a jump for this port is already configured")
						}

						return nil
					}),
				huh.NewInput().
					Title("Interval").
					Description("The rate, per second, a port will change.").
					Placeholder("Number of seconds. Leave blank for default of 30.").
					Value(&intervalString).
					Validate(func(s string) error {
						if s == "" {
							// well default to 30 seconds
							return nil
						}
						p, err := strconv.Atoi(s)
						if err != nil {
							return errors.New("not a valid number")
						}

						if p > 300 {
							return errors.New("the number of seconds chosen is more than 5 minutes")
						}

						return nil
					}),
				huh.NewInput().
					Title("Shared Secret").
					Description("A shared secret to use with HOTP.").
					Placeholder("16 character string. Leave blank to generated one.").
					// CharLimit(16).
					Value(&secret).
					Validate(func(s string) error {
						if s == "" {
							// well default to a generated secret
							return nil
						}

						if len(s) != 16 {
							return fmt.Errorf("enter a string of 16 characters. you entered %d characters", len(s))
						}

						return nil
					}),
				huh.NewConfirm().
					Title("Are you sure you want to add this jump?").
					Affirmative("Yes!").
					Negative("No.").
					Value(&confirm),
			),
		)

		err := form.Run()
		if err == huh.ErrUserAborted {
			return
		}

		if err != nil {
			log.Error().Err(err).Msg("failed to read form input")
			return
		}

		if !confirm {
			fmt.Println("Not adding a new jump.")
			return
		}

		// check for blank values to set defaults
		if intervalString == "" {
			intervalString = "30"
		}

		if secret == "" {
			secret, err = secrets.GenerateTOTPSecret(16)
			if err != nil {
				log.Error().Err(err).Msg("failed to generate a secret")
				return
			}
		}

		port, _ = strconv.Atoi(portString)
		intervalInt, _ := strconv.Atoi(intervalString)
		interval = int64(intervalInt)

		jump, err := options.NewPortJump(port, secret, interval, true)
		if err != nil {
			log.Error().Err(err).Msg("failed to prepare new jump")
			return
		}

		opts.Jumps = append(opts.Jumps, jump)
		if err := opts.Save(); err != nil {
			log.Error().Err(err).Msg("failed to save new jump")
			return
		}

		fmt.Printf("New jump for port %d added!\n", port)
	},
}

func checkIfJumpExists(port int) bool {
	for _, jump := range opts.Jumps {
		if jump.DstPort == port {
			return true
		}
	}

	return false
}

func init() {
	configCmd.AddCommand(addCmd)
}
