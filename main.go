package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

// TODO: refactor everything

var (
	includeSpecial bool
	length         int
	db             *sql.DB
	dbPath         string
	encryptionKey  string
)

var (
	styleHeading = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#1E1E2E")).
			Padding(0, 1)

	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1"))

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8"))

	stylePrompt = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1"))

	stylePassword = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#1E1E2E")).
			Padding(0, 1)
)

func init() {
	rootCmd.Flags().BoolVarP(&includeSpecial, "special", "s", false, "Include special characters")
	rootCmd.Flags().IntVarP(&length, "length", "l", 12, "Password length")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(styleError.Render("Error getting user home directory: " + err.Error()))
		os.Exit(1)
	}

	dbPath = filepath.Join(homeDir, ".fortpass", "passwords.db")
	err = os.MkdirAll(filepath.Dir(dbPath), 0700)
	if err != nil {
		fmt.Println(styleError.Render("Error creating directory: " + err.Error()))
		os.Exit(1)
	}

	// Retrieve encryption key from system keyring
	encryptionKey, err = keyring.Get("fortpass", "db_encryption_key")
	if err != nil {
		// Generate a new encryption key if not found
		encryptionKey = generateEncryptionKey()
		err = keyring.Set("fortpass", "db_encryption_key", encryptionKey)
		if err != nil {
			fmt.Println(styleError.Render("Error storing encryption key: " + err.Error()))
			os.Exit(1)
		}
	}

	// Open the encrypted database
	db, err = sql.Open("sqlite3", fmt.Sprintf("%s?_pragma_key=%s", dbPath, encryptionKey))
	if err != nil {
		fmt.Println(styleError.Render("Error opening database: " + err.Error()))
		os.Exit(1)
	}

	createTable()
}

func generateEncryptionKey() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	key := make([]byte, 32)
	_, err := rng.Read(key)
	if err != nil {
		fmt.Println(styleError.Render("Error generating encryption key: " + err.Error()))
		os.Exit(1)
	}
	return hex.EncodeToString(key)
}

func createTable() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS version (
			version INTEGER PRIMARY KEY
		);
		CREATE TABLE IF NOT EXISTS passwords (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source TEXT,
			username TEXT,
			password TEXT,
			url TEXT,
			UNIQUE(source, username)
		)
	`)
	if err != nil {
		fmt.Println(styleError.Render("Error creating tables: " + err.Error()))
		os.Exit(1)
	}

	// Check and update database version
	checkAndMigrateDatabase()
}

func checkAndMigrateDatabase() {
	var version int
	err := db.QueryRow("SELECT version FROM version").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			// Initialize version
			_, err = db.Exec("INSERT INTO version (version) VALUES (1)")
			if err != nil {
				fmt.Println(styleError.Render("Error initializing database version: " + err.Error()))
				os.Exit(1)
			}
			return
		}
		fmt.Println(styleError.Render("Error checking database version: " + err.Error()))
		os.Exit(1)
	}

	// Implement migrations for future versions
	// if version < 2 {
	//     // Perform migration to version 2
	//     // Update version number
	// }
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

Flags:
  -h, --help   help for fortpass

Use "fortpass [command] --help" for more information about a command.`,
	Run: func(cmd *cobra.Command, args []string) {
		generatePassword()
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all passwords",
	Run: func(cmd *cobra.Command, args []string) {
		showPasswords()
	},
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search passwords",
	Run: func(cmd *cobra.Command, args []string) {
		searchPasswords()
	},
}

var getCmd = &cobra.Command{
	Use:   "get [password name]",
	Short: "Get a specific password and copy it to clipboard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		getPassword(args[0])
	},
}

var importCmd = &cobra.Command{
	Use:   "import [csv_file]",
	Short: "Import passwords from a CSV file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		importPasswords(args[0])
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [source/username]",
	Short: "Delete a specific password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deletePassword(args[0])
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [source/username]",
	Short: "Update a specific password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		updatePassword(args[0])
	},
}

func generatePassword() {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if includeSpecial {
		charset += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rng.Intn(len(charset))]
	}

	fmt.Println(styleHeading.Render("🔐 Generated Password"))
	fmt.Println(stylePassword.Render(string(password)))

	err := clipboard.WriteAll(string(password))
	if err != nil {
		fmt.Println(styleError.Render("❌ Failed to copy password to clipboard: " + err.Error()))
	} else {
		fmt.Println(styleSuccess.Render("📋 Password copied to clipboard."))
	}

	// Call storeInPass with the generated password
	storeInPass(string(password))
}

func showPasswords() {
	passwords := getPasswordEntries()
	fmt.Printf("Debug: Number of passwords found: %d\n", len(passwords))
	if len(passwords) == 0 {
		fmt.Println(stylePrompt.Render("No passwords found in the store."))
		return
	}
	fmt.Println(styleHeading.Render("Available passwords:"))
	for _, entry := range passwords {
		fmt.Printf("%s %s/%s\n", stylePrompt.Render("•"), entry.Source, entry.Username)
	}
}

type PasswordEntry struct {
	Source   string
	Username string
	URL      string
}

func getPasswordEntries() []PasswordEntry {
	rows, err := db.Query("SELECT source, username, url FROM passwords")
	if err != nil {
		fmt.Println(styleError.Render("Error fetching passwords: " + err.Error()))
		return nil
	}
	defer rows.Close()

	var entries []PasswordEntry
	for rows.Next() {
		var entry PasswordEntry
		err := rows.Scan(&entry.Source, &entry.Username, &entry.URL)
		if err != nil {
			fmt.Println(styleError.Render("Error scanning row: " + err.Error()))
			continue
		}
		entries = append(entries, entry)
	}

	return entries
}

func searchPasswords() {
	entries := getPasswordEntries()
	if len(entries) == 0 {
		fmt.Println(stylePrompt.Render("No passwords found in the store."))
		return
	}

	p := tea.NewProgram(initialSearchModel(entries), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	// Handle the selected item
	if m, ok := m.(searchModel); ok && m.selectedItem != nil {
		selectedItem := m.selectedItem.(listItem)
		copyPasswordToClipboard(selectedItem.source, selectedItem.username)
	}
}

type listItem struct {
	source   string
	username string
	url      string
}

func (i listItem) Title() string {
	return fmt.Sprintf("Source: %s | Username: %s", i.source, i.username)
}

func (i listItem) Description() string {
	return fmt.Sprintf("URL: %s", i.url)
}

func (i listItem) FilterValue() string {
	return i.source + i.username + i.url
}

type searchModel struct {
	entries      []PasswordEntry
	searchInput  textinput.Model
	list         list.Model
	selectedItem list.Item
	focused      string // "input" or "list"
}

// Add this method to the searchModel struct
func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func initialSearchModel(entries []PasswordEntry) searchModel {
	m := searchModel{
		entries: entries,
		focused: "input",
	}

	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Search passwords..."
	m.searchInput.Focus()

	items := convertToListItems(entries)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FF00FF")).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder())
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#FF00FF")).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder())

	delegate.SetHeight(3) // Increase height to accommodate two lines

	m.list = list.New(items, delegate, 0, 0)
	m.list.Title = "Passwords"
	m.list.SetStatusBarItemName("password", "passwords")
	m.list.SetSize(100, 20) // Adjust width and height as needed

	m.list.Styles.Title = m.list.Styles.Title.
		Background(lipgloss.Color("#25A065")).
		Foreground(lipgloss.Color("#FFFFFF"))

	return m
}

func convertToListItems(entries []PasswordEntry) []list.Item {
	items := make([]list.Item, len(entries))
	for i, entry := range entries {
		items[i] = listItem{
			source:   entry.Source,
			username: entry.Username,
			url:      entry.URL,
		}
	}
	return items
}

func (m searchModel) View() string {
	var b strings.Builder

	b.WriteString("Search: ")
	if m.focused == "input" {
		b.WriteString(styleSuccess.Render(m.searchInput.View()))
	} else {
		b.WriteString(m.searchInput.View())
	}
	b.WriteString("\n\n")

	if m.focused == "list" {
		b.WriteString(styleSuccess.Render("(Use arrow keys to navigate, Enter to select)\n"))
	}
	b.WriteString(m.list.View())

	return docStyle.Render(b.String())
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.focused == "input" {
				m.focused = "list"
				m.searchInput.Blur()
				return m, nil
			}
			if m.focused == "list" {
				selectedItem := m.list.SelectedItem()
				if selectedItem != nil {
					m.selectedItem = selectedItem
					return m, tea.Quit
				}
			}
		case "esc":
			if m.focused == "list" {
				m.focused = "input"
				m.searchInput.Focus()
			} else {
				return m, tea.Quit
			}
		case "tab":
			if m.focused == "input" {
				m.focused = "list"
				m.searchInput.Blur()
			} else {
				m.focused = "input"
				m.searchInput.Focus()
			}
		case "up", "down":
			if m.focused == "list" {
				var listCmd tea.Cmd
				m.list, listCmd = m.list.Update(msg)
				return m, listCmd
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-3)
	}

	if m.focused == "input" {
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.filterList()
	} else {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m *searchModel) filterList() {
	if m.searchInput.Value() == "" {
		m.list.SetItems(convertToListItems(m.entries))
		return
	}

	pattern := strings.ToLower(m.searchInput.Value())
	var filtered []list.Item
	for _, entry := range m.entries {
		if strings.Contains(strings.ToLower(entry.Source), pattern) ||
			strings.Contains(strings.ToLower(entry.Username), pattern) ||
			strings.Contains(strings.ToLower(entry.URL), pattern) {
			filtered = append(filtered, listItem{
				source:   entry.Source,
				username: entry.Username,
				url:      entry.URL,
			})
		}
	}
	m.list.SetItems(filtered)
}

func getPassword(name string) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		fmt.Println(styleError.Render("❌ Invalid password name format. Use 'source/username'."))
		return
	}

	source, username := parts[0], parts[1]
	var password string
	err := db.QueryRow("SELECT password FROM passwords WHERE source = ? AND username = ?", source, username).Scan(&password)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println(styleError.Render("❌ No password found for " + name))
		} else {
			fmt.Println(styleError.Render("❌ Error fetching password: " + err.Error()))
		}
		return
	}

	err = clipboard.WriteAll(password)
	if err != nil {
		fmt.Println(styleError.Render("❌ Failed to copy password to clipboard: " + err.Error()))
	} else {
		fmt.Printf(styleSuccess.Render("📋 Password for %s/%s copied to clipboard. Will clear in 45 seconds.\n"), source, username)

		go func() {
			time.Sleep(45 * time.Second)
			err := clipboard.WriteAll("")
			if err != nil {
				fmt.Println(styleError.Render("❌ Failed to clear clipboard: " + err.Error()))
			}
		}()
	}
}

type storePasswordModel struct {
	textInputs []textinput.Model
	password   string
	focusIndex int
}

func initialStorePasswordModel(password string) storePasswordModel {
	m := storePasswordModel{
		textInputs: make([]textinput.Model, 3),
		password:   password,
		focusIndex: 0,
	}

	var t textinput.Model
	for i := range m.textInputs {
		t = textinput.New()
		t.CharLimit = 100

		switch i {
		case 0:
			t.Placeholder = "Enter username"
			t.Focus()
		case 1:
			t.Placeholder = "Enter source (e.g., website name, database name)"
		case 2:
			t.Placeholder = "Enter URL"
		}

		m.textInputs[i] = t
	}

	return m
}

func (m storePasswordModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m storePasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.textInputs) {
				return m, tea.Quit
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.textInputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.textInputs)
			}

			cmds := make([]tea.Cmd, len(m.textInputs))
			for i := 0; i <= len(m.textInputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.textInputs[i].Focus()
					continue
				}
				m.textInputs[i].Blur()
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}
func (m *storePasswordModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))

	for i := range m.textInputs {
		m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m storePasswordModel) View() string {
	var b strings.Builder

	for i := range m.textInputs {
		b.WriteString(m.textInputs[i].View())
		if i < len(m.textInputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := "\n[ Store ]"
	if m.focusIndex == len(m.textInputs) {
		button = "\n[ " + styleSuccess.Render("Store") + " ]"
	}
	b.WriteString(button)

	return b.String()
}

func storeInPass(password string) {
	fmt.Println(stylePrompt.Render("Do you want to store this password? (y/n)"))
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println(styleError.Render("❌ Error reading input: " + err.Error()))
		return
	}

	if response != "y" && response != "Y" {
		fmt.Println(stylePrompt.Render("👋 Exiting without storing password."))
		return
	}

	p := tea.NewProgram(initialStorePasswordModel(password))
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	finalModel := m.(storePasswordModel)
	username := finalModel.textInputs[0].Value()
	source := finalModel.textInputs[1].Value()
	url := finalModel.textInputs[2].Value()

	if username == "" || source == "" {
		fmt.Println(stylePrompt.Render("👋 Exiting without storing password."))
		return
	}

	_, err = db.Exec("INSERT INTO passwords (source, username, password, url) VALUES (?, ?, ?, ?)", source, username, password, url)
	if err != nil {
		fmt.Println(styleError.Render("❌ Failed to store password in database: " + err.Error()))
	} else {
		fmt.Println(styleSuccess.Render("✅ Password stored in database successfully."))
	}
}

func copyPasswordToClipboard(source, username string) {
	var password string
	err := db.QueryRow("SELECT password FROM passwords WHERE source = ? AND username = ?", source, username).Scan(&password)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println(styleError.Render(fmt.Sprintf("❌ No password found for %s/%s", source, username)))
		} else {
			fmt.Println(styleError.Render("❌ Error fetching password: " + err.Error()))
		}
		return
	}

	err = clipboard.WriteAll(password)
	if err != nil {
		fmt.Println(styleError.Render("❌ Failed to copy password to clipboard: " + err.Error()))
	} else {
		fmt.Printf(styleSuccess.Render("📋 Password for %s/%s copied to clipboard. Will clear in 45 seconds.\n"), source, username)

		go func() {
			time.Sleep(45 * time.Second)
			err := clipboard.WriteAll("")
			if err != nil {
				fmt.Println(styleError.Render("❌ Failed to clear clipboard: " + err.Error()))
			}
		}()
	}
}

func importPasswords(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(styleError.Render("❌ Error opening CSV file: " + err.Error()))
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println(styleError.Render("❌ Error reading CSV file: " + err.Error()))
		return
	}

	// Skip the header row
	for _, record := range records[1:] {
		source := record[0]
		url := record[1]
		username := record[2]
		password := record[3]

		_, err := db.Exec("INSERT OR REPLACE INTO passwords (source, username, password, url) VALUES (?, ?, ?, ?)", source, username, password, url)
		if err != nil {
			fmt.Printf(styleError.Render("❌ Error importing password for %s: %s\n"), source, err.Error())
		} else {
			fmt.Printf(styleSuccess.Render("✅ Imported password for %s\n"), source)
		}
	}
}

func deletePassword(name string) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		fmt.Println(styleError.Render("❌ Invalid password name format. Use 'source/username'."))
		return
	}

	source, username := parts[0], parts[1]

	result, err := db.Exec("DELETE FROM passwords WHERE source = ? AND username = ?", source, username)
	if err != nil {
		fmt.Println(styleError.Render("❌ Error deleting password: " + err.Error()))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		fmt.Println(styleError.Render(fmt.Sprintf("❌ No password found for %s/%s", source, username)))
	} else {
		fmt.Println(styleSuccess.Render(fmt.Sprintf("✅ Password for %s/%s deleted successfully", source, username)))
	}
}

func updatePassword(name string) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		fmt.Println(styleError.Render("❌ Invalid password name format. Use 'source/username'."))
		return
	}

	source, username := parts[0], parts[1]

	// Check if the password exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM passwords WHERE source = ? AND username = ?)", source, username).Scan(&exists)
	if err != nil {
		fmt.Println(styleError.Render("❌ Error checking password existence: " + err.Error()))
		return
	}

	if !exists {
		fmt.Println(styleError.Render(fmt.Sprintf("❌ No password found for %s/%s", source, username)))
		return
	}

	// Generate a new password
	newPassword := generateNewPassword()

	// Update the password in the database
	_, err = db.Exec("UPDATE passwords SET password = ? WHERE source = ? AND username = ?", newPassword, source, username)
	if err != nil {
		fmt.Println(styleError.Render("❌ Error updating password: " + err.Error()))
		return
	}

	fmt.Println(styleSuccess.Render(fmt.Sprintf("✅ Password for %s/%s updated successfully", source, username)))
	fmt.Println(stylePassword.Render("New password: " + newPassword))

	// Copy the new password to clipboard
	err = clipboard.WriteAll(newPassword)
	if err != nil {
		fmt.Println(styleError.Render("❌ Failed to copy new password to clipboard: " + err.Error()))
	} else {
		fmt.Println(styleSuccess.Render("📋 New password copied to clipboard. Will clear in 45 seconds."))

		go func() {
			time.Sleep(45 * time.Second)
			err := clipboard.WriteAll("")
			if err != nil {
				fmt.Println(styleError.Render("❌ Failed to clear clipboard: " + err.Error()))
			}
		}()
	}
}

func generateNewPassword() string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	password := make([]byte, 16)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}

func main() {
	fmt.Println(styleHeading.Render("🔑 Password Manager CLI"))
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(styleError.Render("Error: " + err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(updateCmd)
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)
