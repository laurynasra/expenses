package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View states
type viewState int

const (
	mainMenuView viewState = iota
	providerSelectionView
	fileInputView
	summaryView
	categoryBrowserView
	uncategorizedView
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginTop(1).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	categoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	amountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)
)

type model struct {
	state               viewState
	selectedProvider    string
	filePath            string
	expenseCategories   *ExpenseCategories
	allExpenses         []*Expense
	unmatchedExpenses   []*Expense
	textInput           textinput.Model
	matcherInput        textinput.Model
	list                list.Model
	cursor              int
	selectedCategoryIdx int
	err                 string
	successMsg          string
	width               int
	height              int
	inputMode           bool // true when entering matcher
}

type item string

func (i item) FilterValue() string { return string(i) }
func (i item) Title() string       { return string(i) }
func (i item) Description() string { return "" }

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter file path..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	mi := textinput.New()
	mi.Placeholder = "Enter keyword to match (e.g., 'amazon', 'netflix')..."
	mi.CharLimit = 100
	mi.Width = 60

	mainMenuItems := []list.Item{
		item("Parse Expenses"),
		item("Exit"),
	}

	l := list.New(mainMenuItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Expense Tracker"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return model{
		state:        mainMenuView,
		textInput:    ti,
		matcherInput: mi,
		list:         l,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-8)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == mainMenuView {
				return m, tea.Quit
			}
			// Go back to main menu from other views
			m.state = mainMenuView
			m.err = ""
			return m, nil

		case "esc":
			// Go back to previous view
			switch m.state {
			case providerSelectionView, summaryView:
				m.state = mainMenuView
			case fileInputView:
				m.state = providerSelectionView
			case categoryBrowserView, uncategorizedView:
				m.state = summaryView
			}
			m.err = ""
			return m, nil
		}

		switch m.state {
		case mainMenuView:
			return m.updateMainMenu(msg)
		case providerSelectionView:
			return m.updateProviderSelection(msg)
		case fileInputView:
			return m.updateFileInput(msg)
		case summaryView:
			return m.updateSummary(msg)
		case categoryBrowserView:
			return m.updateCategoryBrowser(msg)
		case uncategorizedView:
			return m.updateUncategorized(msg)
		}
	}

	return m, nil
}

func (m model) updateMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		selected := m.list.SelectedItem()
		if selected != nil {
			if selected.FilterValue() == "Parse Expenses" {
				// Create provider list
				providers := []list.Item{
					item("wise"),
					item("revolut"),
					item("seb"),
				}
				m.list.SetItems(providers)
				m.list.Title = "Select Provider"
				m.state = providerSelectionView
				return m, nil
			} else if selected.FilterValue() == "Exit" {
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) updateProviderSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		selected := m.list.SelectedItem()
		if selected != nil {
			m.selectedProvider = selected.FilterValue()
			m.state = fileInputView
			m.textInput.SetValue("")
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) updateFileInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		m.filePath = m.textInput.Value()

		// Validate file exists
		absPath, err := filepath.Abs(m.filePath)
		if err != nil {
			m.err = fmt.Sprintf("Invalid path: %v", err)
			return m, nil
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			m.err = fmt.Sprintf("File not found: %s", absPath)
			return m, nil
		}

		// Parse expenses
		categories, expenses, err := parseAndAggregate(m.filePath, m.selectedProvider)
		if err != nil {
			m.err = fmt.Sprintf("Error parsing: %v", err)
			return m, nil
		}

		m.expenseCategories = categories
		m.allExpenses = expenses
		m.unmatchedExpenses = []*Expense{}
		for _, exp := range expenses {
			if !exp.Matched {
				m.unmatchedExpenses = append(m.unmatchedExpenses, exp)
			}
		}

		m.state = summaryView
		m.cursor = 0
		m.err = ""
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) updateSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		maxCursor := len(m.expenseCategories.categories) + 1
		if len(m.unmatchedExpenses) > 0 {
			maxCursor++
		}
		if m.cursor < maxCursor-1 {
			m.cursor++
		}
	case "enter":
		// First N items are categories
		if m.cursor < len(m.expenseCategories.categories) {
			m.selectedCategoryIdx = m.cursor
			m.state = categoryBrowserView
			return m, nil
		}
		// Last item (if exists) is uncategorized
		if len(m.unmatchedExpenses) > 0 && m.cursor == len(m.expenseCategories.categories) {
			m.state = uncategorizedView
			m.selectedCategoryIdx = 0
			return m, nil
		}
	}
	return m, nil
}

func (m model) updateCategoryBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		category := m.expenseCategories.categories[m.selectedCategoryIdx]
		if m.cursor < len(category.Expenses)-1 {
			m.cursor++
		}
	}
	return m, nil
}

func (m model) updateUncategorized(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.unmatchedExpenses) == 0 {
		return m, nil
	}

	// Handle matcher input mode
	if m.inputMode {
		var cmd tea.Cmd
		switch msg.String() {
		case "enter":
			// Get the matcher and save it
			matcher := strings.TrimSpace(m.matcherInput.Value())
			categoryIdx := m.cursor // cursor holds the selected category index during input mode

			if matcher != "" && categoryIdx < len(m.expenseCategories.categories) {
				category := m.expenseCategories.categories[categoryIdx]

				// Add matcher to category and save to file
				category.Matchers = append(category.Matchers, matcher)
				if err := saveCategoriesJSON("categories.json", m.expenseCategories); err != nil {
					m.err = fmt.Sprintf("Failed to save matcher: %v", err)
				}

				// Apply the new matcher to all unmatched expenses
				remainingExpenses := []*Expense{}
				matchedCount := 0
				for _, expense := range m.unmatchedExpenses {
					if strings.Contains(strings.ToLower(expense.Description), strings.ToLower(matcher)) {
						// This expense matches the new matcher - categorize it
						category.Amount += expense.Amount
						expense.Matched = true
						expense.Category = category.Category
						category.Expenses = append(category.Expenses, expense)
						matchedCount++
					} else {
						// Keep in unmatched list
						remainingExpenses = append(remainingExpenses, expense)
					}
				}
				m.unmatchedExpenses = remainingExpenses

				// Set success message
				if matchedCount > 0 {
					m.successMsg = fmt.Sprintf("✓ Categorized %d expense(s) to %s with matcher '%s'",
						matchedCount, category.Category, matcher)
					m.err = ""
				}

				// Adjust cursor to stay in bounds
				if m.selectedCategoryIdx >= len(m.unmatchedExpenses) && len(m.unmatchedExpenses) > 0 {
					m.selectedCategoryIdx = len(m.unmatchedExpenses) - 1
				} else if len(m.unmatchedExpenses) == 0 {
					m.selectedCategoryIdx = 0
				}

				// Exit input mode
				m.inputMode = false
				m.matcherInput.SetValue("")
				m.matcherInput.Blur()

				// Go back to summary if all categorized
				if len(m.unmatchedExpenses) == 0 {
					m.state = summaryView
					m.cursor = 0
				}
			}
			return m, nil

		case "esc":
			// Cancel input mode
			m.inputMode = false
			m.matcherInput.SetValue("")
			m.matcherInput.Blur()
			m.successMsg = "" // Clear success message when canceling
			return m, nil
		}

		m.matcherInput, cmd = m.matcherInput.Update(msg)
		return m, cmd
	}

	// Normal navigation mode
	switch msg.String() {
	case "up", "k":
		if m.selectedCategoryIdx > 0 {
			m.selectedCategoryIdx--
			m.successMsg = "" // Clear success message on navigation
		}
	case "down", "j":
		if m.selectedCategoryIdx < len(m.unmatchedExpenses)-1 {
			m.selectedCategoryIdx++
			m.successMsg = "" // Clear success message on navigation
		}
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Enter input mode for matcher
		categoryIdx := int(msg.String()[0] - '1')
		if categoryIdx < len(m.expenseCategories.categories) {
			m.cursor = categoryIdx
			m.inputMode = true
			m.matcherInput.Focus()
			m.matcherInput.SetValue("")
			m.successMsg = "" // Clear success message when entering input mode
		}
	case "s":
		// Skip this expense without categorizing
		m.unmatchedExpenses = append(m.unmatchedExpenses[:m.selectedCategoryIdx],
			m.unmatchedExpenses[m.selectedCategoryIdx+1:]...)

		if m.selectedCategoryIdx >= len(m.unmatchedExpenses) && m.selectedCategoryIdx > 0 {
			m.selectedCategoryIdx--
		}

		if len(m.unmatchedExpenses) == 0 {
			m.state = summaryView
			m.cursor = 0
		}
		m.successMsg = "" // Clear success message on skip
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.state {
	case mainMenuView:
		return m.viewMainMenu()
	case providerSelectionView:
		return m.viewProviderSelection()
	case fileInputView:
		return m.viewFileInput()
	case summaryView:
		return m.viewSummary()
	case categoryBrowserView:
		return m.viewCategoryBrowser()
	case uncategorizedView:
		return m.viewUncategorized()
	}

	return ""
}

func (m model) viewMainMenu() string {
	return "\n" + m.list.View()
}

func (m model) viewProviderSelection() string {
	help := helpStyle.Render("↑/↓: navigate • enter: select • esc: back • q: quit")
	return "\n" + m.list.View() + "\n" + help
}

func (m model) viewFileInput() string {
	s := titleStyle.Render(fmt.Sprintf("Enter CSV file path (Provider: %s)", m.selectedProvider))
	s += "\n\n" + m.textInput.View() + "\n\n"

	if m.err != "" {
		s += errorStyle.Render("Error: "+m.err) + "\n\n"
	}

	s += helpStyle.Render("enter: parse • esc: back • q: quit")
	return "\n" + s
}

func (m model) viewSummary() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Expense Summary") + "\n\n")

	// Calculate total
	total := 0.0
	for _, cat := range m.expenseCategories.categories {
		total += cat.Amount
	}

	s.WriteString(fmt.Sprintf("Total Expenses: %s\n\n", amountStyle.Render(fmt.Sprintf("€%.2f", total))))

	// Category list with bars
	for i, cat := range m.expenseCategories.categories {
		if cat.Amount == 0 {
			continue
		}

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		// Calculate percentage bar
		percentage := (cat.Amount / total) * 100
		barWidth := int(percentage / 2) // Max 50 chars for 100%
		if barWidth > 50 {
			barWidth = 50
		}
		bar := strings.Repeat("█", barWidth)

		line := fmt.Sprintf("%s %-15s %s %s (%.1f%%)",
			cursor,
			categoryStyle.Render(cat.Category),
			amountStyle.Render(fmt.Sprintf("€%.2f", cat.Amount)),
			bar,
			percentage)

		if m.cursor == i {
			s.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			s.WriteString(line + "\n")
		}
	}

	// Unmatched expenses option
	if len(m.unmatchedExpenses) > 0 {
		s.WriteString("\n")
		cursor := " "
		if m.cursor == len(m.expenseCategories.categories) {
			cursor = ">"
		}
		line := fmt.Sprintf("%s ⚠️  Uncategorized Expenses (%d)", cursor, len(m.unmatchedExpenses))
		if m.cursor == len(m.expenseCategories.categories) {
			s.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			s.WriteString(line + "\n")
		}
	}

	s.WriteString("\n" + helpStyle.Render("↑/↓: navigate • enter: view details • esc: back • q: quit"))

	return "\n" + s.String()
}

func (m model) viewCategoryBrowser() string {
	category := m.expenseCategories.categories[m.selectedCategoryIdx]

	var s strings.Builder
	s.WriteString(titleStyle.Render(fmt.Sprintf("%s - €%.2f", category.Category, category.Amount)) + "\n\n")

	for i, exp := range category.Expenses {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		line := fmt.Sprintf("%s %-50s %s",
			cursor,
			exp.Description,
			amountStyle.Render(fmt.Sprintf("€%.2f", exp.Amount)))

		if m.cursor == i {
			s.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			s.WriteString(line + "\n")
		}
	}

	s.WriteString("\n" + helpStyle.Render("↑/↓: navigate • esc: back • q: quit"))

	return "\n" + s.String()
}

func (m model) viewUncategorized() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(fmt.Sprintf("Uncategorized Expenses (%d)", len(m.unmatchedExpenses))) + "\n\n")

	if len(m.unmatchedExpenses) == 0 {
		s.WriteString("All expenses categorized! 🎉\n\n")
		s.WriteString(helpStyle.Render("esc: back • q: quit"))
		return "\n" + s.String()
	}

	// Show current expense
	exp := m.unmatchedExpenses[m.selectedCategoryIdx]
	expBox := boxStyle.Render(fmt.Sprintf(
		"Description: %s\nAmount: %s\nProvider: %s\n\nExpense %d of %d",
		exp.Description,
		amountStyle.Render(fmt.Sprintf("€%.2f", exp.Amount)),
		exp.Provider,
		m.selectedCategoryIdx+1,
		len(m.unmatchedExpenses),
	))
	s.WriteString(expBox + "\n\n")

	// Show success message if present
	if m.successMsg != "" {
		s.WriteString(successStyle.Render(m.successMsg) + "\n\n")
	}

	// If in input mode, show matcher input
	if m.inputMode {
		selectedCategory := m.expenseCategories.categories[m.cursor]
		s.WriteString(fmt.Sprintf("Categorizing to: %s\n\n", categoryStyle.Render(selectedCategory.Category)))
		s.WriteString("Enter keyword/matcher to save for future auto-categorization:\n")
		s.WriteString(m.matcherInput.View() + "\n\n")

		if m.err != "" {
			s.WriteString(errorStyle.Render("Error: "+m.err) + "\n\n")
		}

		s.WriteString(helpStyle.Render("enter: save & categorize • esc: cancel"))
	} else {
		// Show category options
		s.WriteString("Select category:\n")
		for i, cat := range m.expenseCategories.categories {
			if i < 9 {
				s.WriteString(fmt.Sprintf(" %d. %s\n", i+1, categoryStyle.Render(cat.Category)))
			}
		}

		s.WriteString("\n" + helpStyle.Render("1-9: categorize • s: skip • ↑/↓: navigate expenses • esc: back • q: quit"))
	}

	return "\n" + s.String()
}

func startTUI() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// saveCategoriesJSON saves expense categories back to JSON file
func saveCategoriesJSON(filename string, expenseCategories *ExpenseCategories) error {
	configs := make([]CategoryConfig, 0, len(expenseCategories.categories))

	for _, cat := range expenseCategories.categories {
		configs = append(configs, CategoryConfig{
			Name:     cat.Category,
			Matchers: cat.Matchers,
		})
	}

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal categories: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
