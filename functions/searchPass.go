package functions

import (
	"fmt"
	"os"
	"pw_cli/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var SearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search passwords",
	Run: func(cmd *cobra.Command, args []string) {
		SearchPasswords()
	},
}

func SearchPasswords() {
	entries := utils.GetPasswordEntries()
	if len(entries) == 0 {
		fmt.Println(utils.StylePrompt.Render("No passwords found in the store."))
		return
	}
	p := tea.NewProgram(utils.InitialSearchModel(entries), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	// Handle the selected item
	if m, ok := m.(utils.SearchModel); ok && m.SelectedItem != nil {
		selectedItem := m.SelectedItem.(utils.ListItem)
		utils.CopyPasswordToClipboard(selectedItem.Source, selectedItem.Username)

	}
}
