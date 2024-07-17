package functions

import (
	"database/sql"
	"fmt"
	"pw_cli/utils"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get [password name]",
	Short: "Get a specific password and copy it to clipboard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetPassword(args[0])
	},
}

func GetPassword(name string) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		fmt.Println(utils.StyleError.Render("❌ Invalid password name format. Use 'source/username'."))
		return
	}

	source, username := parts[0], parts[1]
	var password string
	err := utils.DB.QueryRow("SELECT password FROM passwords WHERE source = ? AND username = ?", source, username).Scan(&password)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println(utils.StyleError.Render("❌ No password found for " + name))
		} else {
			fmt.Println(utils.StyleError.Render("❌ Error fetching password: " + err.Error()))
		}
		return
	}

	err = clipboard.WriteAll(password)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Failed to copy password to clipboard: " + err.Error()))
	} else {
		fmt.Printf(utils.StyleSuccess.Render("📋 Password for %s/%s copied to clipboard. Will clear in 45 seconds.\n"), source, username)

		go func() {
			time.Sleep(45 * time.Second)
			err := clipboard.WriteAll("")
			if err != nil {
				fmt.Println(utils.StyleError.Render("❌ Failed to clear clipboard: " + err.Error()))
			}
		}()
	}
}
