package functions

import (
	"fmt"
	"math/rand"
	"pw_cli/utils"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:   "update [source/username]",
	Short: "Update a specific password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		UpdatePassword(args[0])
	},
}

func UpdatePassword(name string) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		fmt.Println(utils.StyleError.Render("❌ Invalid password name format. Use 'source/username'."))
		return
	}

	source, username := parts[0], parts[1]

	// Check if the password exists
	var exists bool
	err := utils.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM passwords WHERE source = ? AND username = ?)", source, username).Scan(&exists)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error checking password existence: " + err.Error()))
		return
	}

	if !exists {
		fmt.Println(utils.StyleError.Render(fmt.Sprintf("❌ No password found for %s/%s", source, username)))
		return
	}

	fmt.Println(utils.StylePrompt.Render("Do you want to generate a new password or input one manually? (g/m):"))
	var choice string
	_, err = fmt.Scanln(&choice)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error reading input: " + err.Error()))
		return
	}

	var newPassword string
	if strings.ToLower(choice) == "g" {
		newPassword = GenerateNewPassword()
		fmt.Println(utils.StylePassword.Render("New generated password: " + newPassword))
	} else {
		fmt.Println(utils.StylePrompt.Render("Enter the new password:"))
		_, err = fmt.Scanln(&newPassword)
		if err != nil {
			fmt.Println(utils.StyleError.Render("❌ Error reading input: " + err.Error()))
			return
		}
	}

	// Update the password in the database
	_, err = utils.DB.Exec("UPDATE passwords SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE source = ? AND username = ?", newPassword, source, username)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Error updating password: " + err.Error()))
		return
	}

	fmt.Println(utils.StyleSuccess.Render(fmt.Sprintf("✅ Password for %s/%s updated successfully", source, username)))
	fmt.Println(utils.StylePassword.Render("New password: " + newPassword))

	// Copy the new password to clipboard
	err = clipboard.WriteAll(newPassword)
	if err != nil {
		fmt.Println(utils.StyleError.Render("❌ Failed to copy new password to clipboard: " + err.Error()))
	} else {
		fmt.Println(utils.StyleSuccess.Render("📋 New password copied to clipboard. Will clear in 45 seconds."))

		go func() {
			time.Sleep(45 * time.Second)
			err := clipboard.WriteAll("")
			if err != nil {
				fmt.Println(utils.StyleError.Render("❌ Failed to clear clipboard: " + err.Error()))
			}
		}()
	}
}

func GenerateNewPassword() string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	password := make([]byte, 16)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}
