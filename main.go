package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	includeSpecial bool
	length         int
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
}

var rootCmd = &cobra.Command{
	Use:   "passgen",
	Short: "A password manager CLI tool",
	Long: `A password manager CLI tool that allows you to generate, store, and retrieve passwords.

Usage:
  passgen [command]

Available Commands:
  generate    Generate a new password
  show        Show all stored passwords
  search      Search for stored passwords
  get         Get a specific password by source/username

Flags:
  -h, --help   help for passgen

Use "passgen [command] --help" for more information about a command.`,
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
}

func getPasswordEntries() []PasswordEntry {
	cmd := exec.Command("pass", "ls")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error fetching passwords: %v\nOutput: %s\n", err, string(output))
		return nil
	}

	var entries []PasswordEntry
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Password Store") || strings.HasSuffix(line, "/") {
			continue
		}
		parts := strings.SplitN(line, "/", 2)
		if len(parts) == 2 {
			entries = append(entries, PasswordEntry{Source: parts[0], Username: strings.TrimSuffix(parts[1], ".gpg")})
		}
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
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// Update the listItem type definition
type listItem struct {
	title string
	desc  string
}

// Implement the FilterValue method for the list.Item interface
func (i listItem) FilterValue() string {
	return i.title
}

// Update the searchModel struct to implement tea.Model
type searchModel struct {
	entries     []PasswordEntry
	searchInput textinput.Model
	list        list.Model
}

// Add the Init method to implement tea.Model interface
func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func initialSearchModel(entries []PasswordEntry) searchModel {
	m := searchModel{
		entries: entries,
	}

	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Search passwords..."
	m.searchInput.Focus()

	m.list = list.New(convertToListItems(entries), list.NewDefaultDelegate(), 0, 0)
	m.list.Title = "Passwords"

	return m
}

// Update the convertToListItems function
func convertToListItems(entries []PasswordEntry) []list.Item {
	listItems := make([]list.Item, len(entries))
	for i, entry := range entries {
		listItems[i] = listItem{
			title: fmt.Sprintf("%s/%s", entry.Source, entry.Username),
			desc:  "Press enter to copy password",
		}
	}
	return listItems
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.searchInput.Focused() {
				searchTerm := strings.ToLower(m.searchInput.Value())
				var filtered []list.Item
				for _, entry := range m.entries {
					if strings.Contains(strings.ToLower(entry.Source), searchTerm) || strings.Contains(strings.ToLower(entry.Username), searchTerm) {
						filtered = append(filtered, listItem{
							title: fmt.Sprintf("%s/%s", entry.Source, entry.Username),
							desc:  "Press enter to copy password",
						})
					}
				}
				m.list.SetItems(filtered)
				m.searchInput.Blur()
				return m, nil
			}
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				getPassword(selectedItem.(listItem).title)
				return m, tea.Quit
			}
		case "tab":
			if m.searchInput.Focused() {
				m.searchInput.Blur()
			} else {
				m.searchInput.Focus()
			}
		case "j", "down":
			m.list.CursorDown()
		case "k", "up":
			m.list.CursorUp()
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.list, _ = m.list.Update(msg)
	return m, cmd
}

func (m searchModel) View() string {
	return docStyle.Render(fmt.Sprintf(
		"%s\n\n%s",
		m.searchInput.View(),
		m.list.View(),
	))
}

func getPassword(name string) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		fmt.Println(styleError.Render("❌ Invalid password name format. Use 'source/username'."))
		return
	}

	source, username := parts[0], parts[1]
	cmd := exec.Command("pass", "show", fmt.Sprintf("%s/%s", source, username))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error fetching password for %s/%s: %v\nOutput: %s\n", source, username, err, string(output))
		return
	}

	// Extract the password from the first line of the output
	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		fmt.Println(styleError.Render("❌ No password found for " + name))
		return
	}
	password := strings.TrimSpace(lines[0])

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

// Add this new type for password storage
type storePasswordModel struct {
	textInputs []textinput.Model
	password   string
	focusIndex int
}

// Add this new function to initialize the store password model
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

// Add these methods for the new model
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

// Modify the storeInPass function
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

	passEntry := fmt.Sprintf("%s\nusername: %s\nsource: %s\nurl: %s", password, username, source, url)
	passName := fmt.Sprintf("%s/%s", source, username)

	cmd := exec.Command("pass", "insert", "-m", passName)
	cmd.Stdin = strings.NewReader(passEntry)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println(styleError.Render("❌ Failed to store password in Pass: " + err.Error()))
	} else {
		fmt.Println(styleSuccess.Render("✅ Password stored in Pass successfully."))
	}
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
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)
