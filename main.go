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
	"github.com/sahilm/fuzzy"
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
	Short: "A password generator CLI tool",
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
	passwords := getPasswords()
	for _, p := range passwords {
		fmt.Printf("Name: %s, Username: %s, Source: %s, URL: %s, Password: %s\n", p.name, p.username, p.source, p.url, p.password)
	}
}

func searchPasswords() {
	p := tea.NewProgram(initialModel(""), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type passwordItem struct {
	name     string
	username string
	source   string
	url      string
	password string
}

func (i passwordItem) Title() string       { return i.name }
func (i passwordItem) Description() string { return i.source }
func (i passwordItem) FilterValue() string { return i.name + i.username + i.source + i.url }

type model struct {
	textInputs   []textinput.Model
	password     string
	focusIndex   int
	passwordList list.Model
	searchInput  textinput.Model
	passwords    []passwordItem
}

func initialModel(password string) model {
	m := model{
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

	m.textInputs[0].Focus()
	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Search passwords..."
	m.searchInput.Focus()

	m.passwords = getPasswords()
	m.passwordList = list.New(convertToListItems(m.passwords), list.NewDefaultDelegate(), 0, 0)
	m.passwordList.Title = "Passwords"

	return m
}

func getPasswords() []passwordItem {
	cmd := exec.Command("pass", "grep", "-l", ".")
	output, err := cmd.CombinedOutput() // Capture both stdout and stderr
	if err != nil {
		fmt.Printf("Error fetching passwords: %v\nOutput: %s\n", err, string(output))
		return nil
	}

	var passwords []passwordItem
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line != "" {
			name := strings.TrimSpace(line)
			passwords = append(passwords, getPasswordDetails(name))
		}
	}
	return passwords
}

func getPasswordDetails(name string) passwordItem {
	cmd := exec.Command("pass", "show", name)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error fetching password details for %s: %v\n", name, err)
		return passwordItem{name: name}
	}

	lines := strings.Split(string(output), "\n")
	item := passwordItem{name: name, password: lines[0]}
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, "username:") {
			item.username = strings.TrimSpace(strings.TrimPrefix(line, "username:"))
		} else if strings.HasPrefix(line, "source:") {
			item.source = strings.TrimSpace(strings.TrimPrefix(line, "source:"))
		} else if strings.HasPrefix(line, "url:") {
			item.url = strings.TrimSpace(strings.TrimPrefix(line, "url:"))
		}
	}
	return item
}

func convertToListItems(passwords []passwordItem) []list.Item {
	items := make([]list.Item, len(passwords))
	for i, p := range passwords {
		items[i] = p
	}
	return items
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.searchInput.Focused() {
				m.passwordList.SetItems(m.filterPasswords(m.searchInput.Value()))
				m.searchInput.Blur()
				return m, nil
			}
			selectedItem := m.passwordList.SelectedItem()
			if selectedItem != nil {
				password := selectedItem.(passwordItem)
				err := clipboard.WriteAll(fmt.Sprintf("Password: %s\nUsername: %s\nSource: %s\nURL: %s", password.password, password.username, password.source, password.url))
				if err != nil {
					fmt.Println("Error copying to clipboard:", err)
				}
				return m, tea.Quit
			}
		case "tab":
			if m.searchInput.Focused() {
				m.searchInput.Blur()
			} else {
				m.searchInput.Focus()
			}
		case "j", "down":
			m.passwordList.CursorDown()
		case "k", "up":
			m.passwordList.CursorUp()
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.passwordList.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.passwordList, _ = m.passwordList.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(fmt.Sprintf(
		"%s\n\n%s",
		m.searchInput.View(),
		m.passwordList.View(),
	))
}

func (m *model) filterPasswords(query string) []list.Item {
	if query == "" {
		return convertToListItems(m.passwords)
	}

	// Extract the filter values from password items
	var filterValues []string
	for _, p := range m.passwords {
		filterValues = append(filterValues, p.FilterValue())
	}

	// Perform fuzzy search on the extracted filter values
	matches := fuzzy.Find(query, filterValues)

	// Collect the matched password items
	var filtered []list.Item
	for _, match := range matches {
		filtered = append(filtered, m.passwords[match.Index])
	}
	return filtered
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

	p := tea.NewProgram(initialModel(password))
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	finalModel := m.(model)
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
	fmt.Println(styleHeading.Render("🔑 Password Generator CLI"))
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(styleError.Render("Error: " + err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(searchCmd)
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)
