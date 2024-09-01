package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a jump",
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
					Title("Are you sure you want to delete this jump?").
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
			fmt.Println("Not deleting a jump.")
			return
		}

		if err := deleteJump((selectedJumpPort)); err != nil {
			log.Error().Err(err).Msg("failed to delete jump")
			return
		}

		fmt.Printf("Jump %d deleted.\n", selectedJumpPort)
	},
}

func deleteJump(port int) error {
	index := -1

	// Find the index of the jump to delete
	for i, jump := range opts.Jumps {
		if jump.DstPort == port {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("jump with port %d not found", port)
	}

	// Remove the jump from the slice
	opts.Jumps = append(opts.Jumps[:index], opts.Jumps[index+1:]...)
	if err := opts.Save(); err != nil {
		return err
	}

	return nil
}

func init() {
	configCmd.AddCommand(deleteCmd)
}
