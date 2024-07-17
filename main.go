package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	includeSpecial bool
	length         int
)

func init() {
	rootCmd.Flags().BoolVarP(&includeSpecial, "special", "s", false, "Include special characters")
	rootCmd.Flags().IntVarP(&length, "length", "l", 12, "Password length")
}

var rootCmd = &cobra.Command{
	Use:   "passgen",
	Short: "A password generator CLI tool",
	Run: func(cmd *cobra.Command, args []string) {
		generatePassword()
	},
}

func generatePassword() {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if includeSpecial {
		charset += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}

	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}

	color.Cyan("🔐 Generated password: ")
	color.Green("%s\n", string(password))

	err := clipboard.WriteAll(string(password))
	if err != nil {
		color.Red("❌ Failed to copy password to clipboard: %v\n", err)
	} else {
		color.Yellow("📋 Password copied to clipboard.\n")
	}

	storeInPass(string(password))
}

func storeInPass(password string) {
	color.Cyan("💾 Do you want to store the password in Pass? (y/n): ")
	var answer string
	fmt.Scanln(&answer)

	if strings.ToLower(answer) == "y" {
		var username, source string
		color.Cyan("👤 Enter username: ")
		fmt.Scanln(&username)
		color.Cyan("🌐 Enter source (URL, database, etc.): ")
		fmt.Scanln(&source)

		passEntry := fmt.Sprintf("%s\nusername: %s\nsource: %s", password, username, source)
		passName := fmt.Sprintf("%s/%s", source, username)

		cmd := exec.Command("pass", "insert", "-m", passName)
		cmd.Stdin = strings.NewReader(passEntry)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			color.Red("❌ Failed to store password in Pass: %v\n", err)
		} else {
			color.Green("✅ Password stored in Pass successfully.\n")
		}
	}

	color.Cyan("👋 Exiting.")
}

func main() {
	color.Blue("🔑 Password Generator CLI\n")
	if err := rootCmd.Execute(); err != nil {
		color.Red("Error: %v\n", err)
		os.Exit(1)
	}
}
