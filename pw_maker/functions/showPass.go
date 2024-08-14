package functions

import (
	"fmt"

	"github.com/tadeasf/pw_maker/utils"

	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all passwords",
	Run: func(cmd *cobra.Command, args []string) {
		ShowPasswords()
	},
}

func ShowPasswords() {
	passwords := utils.GetPasswordEntries()
	fmt.Printf("Debug: Number of passwords found: %d\n", len(passwords))
	if len(passwords) == 0 {
		fmt.Println(utils.StylePrompt.Render("No passwords found in the store."))
		return
	}
	fmt.Println(utils.StyleHeading.Render("Available passwords:"))
	for _, entry := range passwords {
		fmt.Printf("%s %s/%s\n", utils.StylePrompt.Render("•"), entry.Source, entry.Username)
	}
}
