package functions

import (
	"encoding/csv"
	"fmt"
	"os"
	"pw_cli/utils"

	"github.com/spf13/cobra"
)

var ImportCmd = &cobra.Command{
	Use:   "import [csv_file]",
	Short: "Import passwords from a CSV file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ImportPasswords(args[0])
	},
}

func ImportPasswords(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error opening CSV file: " + err.Error()))
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error reading CSV file: " + err.Error()))
		return
	}

	tx, err := utils.DB.Begin()
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error starting transaction: " + err.Error()))
		return
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO passwords (source, username, password, url, created_at, updated_at)
		VALUES (?, ?, ?, ?, COALESCE((SELECT created_at FROM passwords WHERE source = ? AND username = ? AND url = ?), CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)
	`)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error preparing statement: " + err.Error()))
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			fmt.Println(utils.StyleError.Render("❌ Error rolling back transaction: " + rollbackErr.Error()))
		}
		return
	}
	defer stmt.Close()

	for _, record := range records[1:] {
		source := record[0]
		url := utils.BeautifyURL(record[1])
		username := record[2]
		password := record[3]

		_, err := stmt.Exec(source, username, password, url, source, username, url)
		if err != nil {
			fmt.Printf(utils.StyleError.Render("❌ Error importing password for %s: %s\n"), source, err.Error())
		} else {
			fmt.Printf(utils.StyleSuccess.Render("✅ Imported/Updated password for %s\n"), source)
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error committing transaction: " + err.Error()))
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			fmt.Println(utils.StyleError.Render("❌ Error rolling back transaction: " + rollbackErr.Error()))
		}
	}
}
