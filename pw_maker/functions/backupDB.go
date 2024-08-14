package functions

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tadeasf/pw_maker/pw_maker/utils"

	"github.com/spf13/cobra"
)

var BackupCmd = &cobra.Command{
	Use:   "backup [destination]",
	Short: "Backup the password database",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var destination string
		if len(args) > 0 {
			destination = args[0]
		} else {
			destination = fmt.Sprintf("fortpass_backup_%s.db", time.Now().Format("20060102_150405"))
		}
		BackupDatabase(destination)
	},
}

func BackupDatabase(destination string) {
	// Close the current database connection
	utils.DB.Close()

	// Copy the database file
	src, err := os.Open(utils.DBPath)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error opening source database: " + err.Error()))
		return
	}
	defer src.Close()

	dst, err := os.Create(destination)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error creating backup file: " + err.Error()))
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error copying database: " + err.Error()))
		return
	}

	fmt.Println(utils.StyleSuccess.Render("✅ Database backed up successfully to: " + destination))

	// Reopen the database connection
	utils.InitDB()
}
