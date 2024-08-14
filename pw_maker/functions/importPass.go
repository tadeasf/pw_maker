package functions

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/tadeasf/pw_maker/pw_maker/utils"

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
		INSERT INTO passwords (source, username, password, url, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(source, username, url) DO UPDATE SET
		password = excluded.password,
		updated_at = CURRENT_TIMESTAMP
		WHERE password != excluded.password
	`)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error preparing statement: " + err.Error()))
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			fmt.Println(utils.StyleError.Render("❌ Error rolling back transaction: " + rollbackErr.Error()))
		}
		return
	}
	defer stmt.Close()

	var inserted, updated, skipped int
	for _, record := range records[1:] {
		source := record[0]
		url := utils.BeautifyURL(record[1])
		username := record[2]
		password := record[3]

		result, err := stmt.Exec(source, username, password, url)
		if err != nil {
			fmt.Printf(utils.StyleError.Render("❌ Error importing password for %s: %s\n"), source, err.Error())
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			skipped++
			fmt.Printf(utils.StyleInfo.Render("ℹ️ Skipped duplicate password for %s (no changes)\n"), source)
		} else if rowsAffected == 1 {
			inserted++
			fmt.Printf(utils.StyleSuccess.Render("✅ Imported new password for %s\n"), source)
		} else {
			updated++
			fmt.Printf(utils.StyleSuccess.Render("✅ Updated existing password for %s\n"), source)
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error committing transaction: " + err.Error()))
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			fmt.Println(utils.StyleError.Render("❌ Error rolling back transaction: " + rollbackErr.Error()))
		}
	} else {
		fmt.Printf(utils.StyleSuccess.Render("Import completed: %d inserted, %d updated, %d skipped\n"), inserted, updated, skipped)
	}
}
