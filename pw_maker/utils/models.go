package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PasswordEntry struct {
	Source    string
	Username  string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListItem struct {
	Source    string
	Username  string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (i ListItem) Title() string {
	return fmt.Sprintf("Source: %s | Username: %s", i.Source, i.Username)
}

func (i ListItem) Description() string {
	return fmt.Sprintf("URL: %s | Created: %s | Updated: %s", i.URL, i.CreatedAt.Format("2006-01-02 15:04:05"), i.UpdatedAt.Format("2006-01-02 15:04:05"))
}

func (i ListItem) FilterValue() string {
	return i.Source + i.Username + i.URL
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func ConvertToListItems(entries []PasswordEntry) []list.Item {
	items := make([]list.Item, len(entries))
	for i, entry := range entries {
		items[i] = ListItem(entry)
	}
	return items
}

type SearchModel struct {
	entries      []PasswordEntry
	searchInput  textinput.Model
	list         list.Model
	SelectedItem list.Item
	focused      string // "input" or "list"
}

// Add this method to the searchModel struct
func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

type StorePasswordModel struct {
	textInputs []textinput.Model
	password   string
	focusIndex int
}

func initialStorePasswordModel(password string) StorePasswordModel {
	m := StorePasswordModel{
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

func (m StorePasswordModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m StorePasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *StorePasswordModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))

	for i := range m.textInputs {
		m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m StorePasswordModel) View() string {
	var b strings.Builder

	for i := range m.textInputs {
		b.WriteString(m.textInputs[i].View())
		if i < len(m.textInputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := "\n[ Store ]"
	if m.focusIndex == len(m.textInputs) {
		button = "\n[ " + StyleSuccess.Render("Store") + " ]"
	}
	b.WriteString(button)

	return b.String()
}

func InitialSearchModel(entries []PasswordEntry) SearchModel {
	m := SearchModel{
		entries: entries,
		focused: "input",
	}

	m.searchInput = textinput.New()
	m.searchInput.Placeholder = "Search passwords..."
	m.searchInput.Focus()

	items := ConvertToListItems(entries)

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

func (m SearchModel) View() string {
	var b strings.Builder

	b.WriteString("Search: ")
	if m.focused == "input" {
		b.WriteString(StyleSuccess.Render(m.searchInput.View()))
	} else {
		b.WriteString(m.searchInput.View())
	}
	b.WriteString("\n\n")

	if m.focused == "list" {
		b.WriteString(StyleSuccess.Render("(Use arrow keys to navigate, Enter to select)\n"))
	}

	b.WriteString(m.list.View())

	return docStyle.Render(b.String())
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					m.SelectedItem = selectedItem
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

func (m *SearchModel) filterList() {
	if m.searchInput.Value() == "" {
		m.list.SetItems(ConvertToListItems(m.entries))
		return
	}

	pattern := strings.ToLower(m.searchInput.Value())
	var filtered []list.Item
	for _, entry := range m.entries {
		if strings.Contains(strings.ToLower(entry.Source), pattern) ||
			strings.Contains(strings.ToLower(entry.Username), pattern) ||
			strings.Contains(strings.ToLower(entry.URL), pattern) {
			filtered = append(filtered, ListItem(entry))
		}
	}
	m.list.SetItems(filtered)
}
