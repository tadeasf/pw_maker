package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
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

	storeInPass(string(password))
}

type model struct {
	textInputs []textinput.Model
	password   string
	focusIndex int
}

func initialModel(password string) model {
	m := model{
		textInputs: make([]textinput.Model, 2),
		password:   password,
		focusIndex: 0,
	}

	var t textinput.Model
	for i := range m.textInputs {
		t = textinput.New()
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Enter username"
			t.Focus()
		case 1:
			t.Placeholder = "Enter source (URL, database, etc.)"
		}

		m.textInputs[i] = t
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))

	for i := range m.textInputs {
		m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(styleHeading.Render("Store Password in Pass"))
	b.WriteString("\n\n")

	for i := range m.textInputs {
		b.WriteString(m.textInputs[i].View())
		b.WriteString("\n")
	}

	button := "[ Store ]"
	if m.focusIndex == len(m.textInputs) {
		button = stylePassword.Render(button)
	}
	fmt.Fprintf(&b, "\n%s\n", button)

	b.WriteString("\n(tab to navigate • enter to select)")

	return b.String()
}

func storeInPass(password string) {
	p := tea.NewProgram(initialModel(password))
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	finalModel := m.(model)
	username := finalModel.textInputs[0].Value()
	source := finalModel.textInputs[1].Value()

	if username == "" || source == "" {
		fmt.Println(stylePrompt.Render("👋 Exiting without storing password."))
		return
	}

	passEntry := fmt.Sprintf("%s\nusername: %s\nsource: %s", password, username, source)
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
