package functions

import (
	"fmt"
	"io"
	"os"
	"pw_cli/utils"

	"github.com/spf13/cobra"
)

var ImportDBCmd = &cobra.Command{
	Use:   "importdb [source]",
	Short: "Import a password database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ImportDatabase(args[0])
	},
}

func ImportDatabase(dbPath string) {
	// Check if the file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println(utils.StyleError.Render("Error: The specified database file does not exist."))
		return
	}
	// Close the current database connection
	utils.DB.Close()

	// Backup the current database before importing
	backupPath := utils.DBPath + ".bak"
	err := os.Rename(utils.DBPath, backupPath)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error backing up current database: " + err.Error()))
		return
	}

	// Copy the imported database file
	src, err := os.Open(dbPath)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error opening source database: " + err.Error()))
		err = os.Rename(backupPath, utils.DBPath) // Restore the backup
		if err != nil {
			fmt.Println(utils.StyleError.Render("❌ Error restoring backup: " + err.Error()))
			return
		}
		return
	}
	defer src.Close()

	dst, err := os.Create(utils.DBPath)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error creating destination database: " + err.Error()))
		err = os.Rename(backupPath, utils.DBPath) // Restore the backup
		if err != nil {
			fmt.Println(utils.StyleError.Render("❌ Error restoring backup: " + err.Error()))
			return
		}
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error copying database: " + err.Error()))
		err = os.Rename(backupPath, utils.DBPath) // Restore the backup
		if err != nil {
			fmt.Println(utils.StyleError.Render("❌ Error restoring backup: " + err.Error()))
			return
		}
		return
	}

	fmt.Println(utils.StyleSuccess.Render("✅ Database imported successfully from: " + dbPath))

	// Reopen the database connection
	utils.InitDB()
}
