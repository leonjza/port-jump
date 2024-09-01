package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// toggleCmd represents the toggle command
var toggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle jump status",
	Run: func(cmd *cobra.Command, args []string) {
		if len(opts.Jumps) == 0 {
			fmt.Println("There are no configured jumps to toggle.")
			return
		}

		selectOptions := make([]huh.Option[int], 0, len(opts.Jumps))
		for _, jump := range opts.Jumps {
			selectOptions = append(selectOptions, huh.NewOption(
				fmt.Sprintf("Port %d, Interval %d (enabled: %s)", jump.DstPort, jump.Interval, styledBool(jump.Enabled)),
				jump.DstPort,
			))
		}

		var selectedJumpPort int
		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select a jump").
					Options(selectOptions...).
					Value(&selectedJumpPort),
				huh.NewConfirm().
					Title("Are you sure you want to toggle this jump?").
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
			fmt.Println("Not toggling a jump.")
			return
		}

		var newState bool
		for _, jump := range opts.Jumps {
			if jump.DstPort == selectedJumpPort {
				jump.Enabled = !jump.Enabled
				newState = jump.Enabled
			}
		}

		if err := opts.Save(); err != nil {
			log.Error().Err(err).Msg("failed to save jump configuration")
			return
		}

		fmt.Printf("Jump %d toggled to %s.\n", selectedJumpPort, styledBool(newState))
	},
}

func init() {
	configCmd.AddCommand(toggleCmd)
}
