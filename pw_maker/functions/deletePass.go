// Func: DeletePassword(name string)
// DeletePassword deletes a password from the database
// It takes a string in the format of "source/username" and deletes the password from the database
package functions

import (
	"fmt"
	"strings"

	"github.com/tadeasf/pw_maker/pw_maker/utils"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [source/username]",
	Short: "Delete a specific password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		DeletePassword(args[0])
	},
}

func DeletePassword(name string) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		fmt.Println(utils.StyleError.Render("❌ Invalid password name format. Use 'source/username'."))
		return
	}

	source, username := parts[0], parts[1]

	result, err := utils.DB.Exec("DELETE FROM passwords WHERE source = ? AND username = ?", source, username)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error deleting password: " + err.Error()))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		fmt.Println(utils.StyleError.Render(fmt.Sprintf("❌ No password found for %s/%s", source, username)))
	} else {
		fmt.Println(utils.StyleSuccess.Render(fmt.Sprintf("✅ Password for %s/%s deleted successfully", source, username)))
	}
}
