package utils

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	StyleHeading = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#1E1E2E")).
			Padding(0, 1)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1"))
	StyleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8"))

	StyleInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1"))
	StylePrompt = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1"))

	StylePassword = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#1E1E2E")).
			Padding(0, 1)
)

var DocStyle = lipgloss.NewStyle().Margin(1, 2)

func BeautifyURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsedURL.String()
}

func CopyPasswordToClipboard(source, username string) {
	var password string
	err := DB.QueryRow("SELECT password FROM passwords WHERE source = ? AND username = ?", source, username).Scan(&password)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println(StyleError.Render(fmt.Sprintf("❌ No password found for %s/%s", source, username)))
		} else {
			fmt.Println(StyleError.Render("❌ Error fetching password: " + err.Error()))
		}
		return
	}
	err = clipboard.WriteAll(password)
	if err != nil {
		fmt.Println(StyleError.Render("❌ Failed to copy password to clipboard: " + err.Error()))
	} else {
		fmt.Printf(StyleSuccess.Render("📋 Password for %s/%s copied to clipboard. Will clear in 45 seconds.\n"), source, username)

		go func() {
			time.Sleep(45 * time.Second)
			err := clipboard.WriteAll("")
			if err != nil {
				fmt.Println(StyleError.Render("❌ Failed to clear clipboard: " + err.Error()))
			}
		}()
	}
}

func GeneratePassword() {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if IncludeSpecial {
		charset += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	password := make([]byte, Length)
	for i := range password {
		password[i] = charset[rng.Intn(len(charset))]
	}

	fmt.Println(StyleHeading.Render("🔐 Generated Password"))
	fmt.Println(StylePassword.Render(string(password)))

	err := clipboard.WriteAll(string(password))
	if err != nil {
		fmt.Println(StyleError.Render("❌ Failed to copy password to clipboard: " + err.Error()))
	} else {
		fmt.Println(StyleSuccess.Render("📋 Password copied to clipboard."))
	}

	// Call storeInPass with the generated password
	storeInPass(string(password))
}
func storeInPass(password string) {
	fmt.Println(StylePrompt.Render("Do you want to store this password? (y/n)"))
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println(StyleError.Render("❌ Error reading input: " + err.Error()))
		return
	}

	if response != "y" && response != "Y" {
		fmt.Println(StylePrompt.Render("👋 Exiting without storing password."))
		return
	}

	p := tea.NewProgram(initialStorePasswordModel(password))
	m, err := p.Run()
	if err != nil {
		fmt.Println(StyleError.Render("Error running program: " + err.Error()))
		os.Exit(1)
	}

	finalModel := m.(StorePasswordModel)
	username := finalModel.textInputs[0].Value()
	source := finalModel.textInputs[1].Value()
	url := BeautifyURL(finalModel.textInputs[2].Value())

	if username == "" || source == "" {
		fmt.Println(StylePrompt.Render("👋 Exiting without storing password."))
		return
	}

	_, err = DB.Exec(`
		INSERT OR REPLACE INTO passwords (source, username, password, url, created_at, updated_at)
		VALUES (?, ?, ?, ?, COALESCE((SELECT created_at FROM passwords WHERE source = ? AND username = ? AND url = ?), CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)
	`, source, username, password, url, source, username, url)
	if err != nil {
		fmt.Println(StyleError.Render("❌ Failed to store password in database: " + err.Error()))
	} else {
		fmt.Println(StyleSuccess.Render("✅ Password stored/updated in database successfully."))
	}
}
