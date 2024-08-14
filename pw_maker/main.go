package main

import (
	"fmt"
	"os"

	"github.com/tadeasf/pw_maker/functions"
	"github.com/tadeasf/pw_maker/utils"

	"github.com/spf13/cobra"
)

// TODO: refactor everything
func init() {
	rootCmd.Flags().BoolVarP(&utils.IncludeSpecial, "special", "s", false, "Include special characters")
	rootCmd.Flags().IntVarP(&utils.Length, "length", "l", 12, "Password length")

	utils.InitDB()

	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(backupDBCmd)
	rootCmd.AddCommand(importDBCmd)
}

var rootCmd = &cobra.Command{
	Use:   "fortpass",
	Short: "A password manager CLI tool",
	Long: `A password manager CLI tool that allows you to generate, store, and retrieve passwords.

Usage:
  fortpass [command]

Available Commands:
  generate    Generate a new password
  show        Show all stored passwords
  search      Search for stored passwords
  get         Get a specific password by source/username
  import      Import passwords from a CSV file
  delete      Delete a specific password
  update      Update a specific password
  backup      Backup the password database
  importdb    Import a password database

Flags:
  -h, --help   help for fortpass

Use "fortpass [command] --help" for more information about a command.`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.GeneratePassword()
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all passwords",
	Run: func(cmd *cobra.Command, args []string) {
		functions.ShowPasswords()
	},
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search passwords",
	Run: func(cmd *cobra.Command, args []string) {
		functions.SearchPasswords()
	},
}

var getCmd = &cobra.Command{
	Use:   "get [password name]",
	Short: "Get a specific password and copy it to clipboard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functions.GetPassword(args[0])
	},
}

var importCmd = &cobra.Command{
	Use:   "import [csv_file]",
	Short: "Import passwords from a CSV file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functions.ImportPasswords(args[0])
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [source/username]",
	Short: "Delete a specific password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functions.DeletePassword(args[0])
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [source/username]",
	Short: "Update a specific password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functions.UpdatePassword(args[0])
	},
}

var backupDBCmd = &cobra.Command{
	Use:   "backupdb [destination]",
	Short: "Backup the password database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functions.BackupDatabase(args[0])
	},
}

var importDBCmd = &cobra.Command{
	Use:   "importdb [db_file]",
	Short: "Import a password database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functions.ImportDatabase(args[0])
	},
}

func main() {
	fmt.Println(utils.StyleHeading.Render("🔑 Password Manager CLI"))
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(utils.StyleError.Render("Error: " + err.Error()))
		os.Exit(1)
	}
}
