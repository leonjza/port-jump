package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the current jumps",
	Run: func(cmd *cobra.Command, args []string) {

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
			StyleFunc(func(row, col int) lipgloss.Style {
				switch {
				case row == 0:
					return lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				default:
					return lipgloss.NewStyle().Padding(0, 1)
				}
			}).
			Headers("Enabled", "Destination", "Interval")

		for _, jump := range opts.Jumps {
			t.Row(
				styledBool(jump.Enabled),
				fmt.Sprintf("%d", jump.DstPort),
				fmt.Sprintf("%d", jump.Interval),
			)
		}

		footer := lipgloss.NewStyle().Bold(true).
			Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("Total jumps: %d", len(opts.Jumps)))

		fmt.Println(t.Render() + "\n" + footer)
	},
}

func styledBool(value bool) string {
	var (
		trueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // Green
		falseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	)

	if value {
		return trueStyle.Render("true")
	}

	return falseStyle.Render("false")
}

func init() {
	configCmd.AddCommand(listCmd)
}
