package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Expense struct {
	Amount      float64
	Description string
	Date        time.Time
	Provider    string
	Category    string
	Matched     bool
}

type ExpenseCategory struct {
	Amount   float64
	Category string
	Expenses []*Expense
	Matchers []string
}

type ExpenseCategories struct {
	categories []*ExpenseCategory
}

// CategoryConfig represents the JSON structure for categories
type CategoryConfig struct {
	Name     string   `json:"name"`
	Matchers []string `json:"matchers"`
}

// loadCategoriesFromJSON loads expense categories from a JSON file
func loadCategoriesFromJSON(filename string) (*ExpenseCategories, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open categories file: %w", err)
	}
	defer file.Close()

	var configs []CategoryConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configs); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	expenseCategories := &ExpenseCategories{}
	for _, config := range configs {
		expenseCategories.AddCategory(&ExpenseCategory{
			Amount:   0,
			Category: config.Name,
			Matchers: config.Matchers,
		})
	}

	return expenseCategories, nil
}

func (e *ExpenseCategories) AddCategory(expense *ExpenseCategory) {
	e.categories = append(e.categories, expense)
}

func (e *ExpenseCategory) Match(description string) bool {
	for _, matcher := range e.Matchers {
		if strings.Contains(description, matcher) {
			return true
		}
	}
	return false
}

func MapWiseExpense(row map[string]string) (*Expense, error) {
	amount, err := strconv.ParseFloat(row["Amount"], 64)
	if err != nil {
		return nil, err
	}
	amount = amount * -1 // Wise shows positive amounts for debits

	description := row["Description"]

	return &Expense{
		Amount:      amount,
		Description: description,
		Provider:    "Wise",
	}, nil
}

func mapRevolutExpense(row map[string]string) (*Expense, error) {
	amount, err := strconv.ParseFloat(row["Amount"], 64)
	if err != nil {
		return nil, err
	}
	description := row["Description"]
	return &Expense{
		Amount:      amount,
		Description: description,
		Provider:    "Revolut",
	}, nil
}

func mapSEBExpense(row map[string]string) (*Expense, error) {
	amount, err := strconv.ParseFloat(strings.ReplaceAll(row["SUMA"], ",", "."), 64)
	if err != nil {
		return nil, err
	}
	description := strings.Join([]string{row["MOKĖTOJO ARBA GAVĖJO PAVADINIMAS"], row["MOKĖJIMO PASKIRTIS"], row["TRANSAKCIJOS TIPAS"]}, " ")

	return &Expense{
		Amount:      amount,
		Description: description,
		Provider:    "SEB",
	}, nil
}

func parseAndAggregate(fileName string, provider string) (*ExpenseCategories, []*Expense, error) {
	// Load categories from JSON file
	expenseCategories, err := loadCategoriesFromJSON("categories.json")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load categories: %w", err)
	}

	parsedExpenses, err := parseExpenses(fileName, provider)
	if err != nil {
		return nil, nil, err
	}

	for _, expense := range parsedExpenses {
		for _, expenseCategory := range expenseCategories.categories {
			if expenseCategory.Match(strings.ToLower(expense.Description)) {
				expenseCategory.Amount += expense.Amount
				expense.Matched = true
				expenseCategory.Expenses = append(expenseCategory.Expenses, expense)
				break //stop matching further categories
			}
		}
	}
	unmatchedExpenses := []*Expense{}
	for _, expense := range parsedExpenses {
		if !expense.Matched {
			unmatchedExpenses = append(unmatchedExpenses, expense)
		}
	}
	return expenseCategories, parsedExpenses, nil
}

func main() {
	//parseAndAggregate("./Išrašas (1).csv", "seb")

	fmt.Println("Starting Expense Tracker...")
	fmt.Println("Creating tview application...")

	app := tview.NewApplication()
	if app == nil {
		fmt.Println("Error: Failed to create tview application")
		return
	}

	fmt.Println("Application created successfully")
	fmt.Println("Setting up main menu...")

	// Start with the main menu
	showMainMenu(app)

	fmt.Println("Starting TUI...")
	fmt.Println("If you see this message but no menu, there might be a terminal compatibility issue.")
	fmt.Println("Try resizing your terminal or pressing Enter/Escape.")

	// Run the application
	if err := app.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		log.Fatal(err)
	}

	fmt.Println("Application exited cleanly.")
}

func showSupportedProviders(app *tview.Application) {
	supportedProviders := []string{"Wise", "SEB", "Revolut"}

	// Simple, clear text
	textContent := "Supported Providers:\n\n"
	for i, provider := range supportedProviders {
		textContent += fmt.Sprintf("%d. %s\n", i+1, provider)
	}
	textContent += "\nPress ESC to go back or 'q' to quit"

	text := tview.NewTextView()
	text.SetText(textContent)
	text.SetBorder(true)
	text.SetTitle("Supported Providers")

	text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			showMainMenu(app)
		case tcell.KeyRune:
			if event.Rune() == 'q' || event.Rune() == 'Q' {
				app.Stop()
			}
		}
		return event
	})

	app.SetRoot(text, true)
}

func showParseExpensesForm(app *tview.Application) {
	var form *tview.Form

	form = tview.NewForm().
		AddInputField("File Path", "", 50, nil, nil).
		AddDropDown("Provider", []string{"wise", "seb", "revolut"}, 0, nil).
		AddButton("Parse", func() {
			filePath := form.GetFormItem(0).(*tview.InputField).GetText()
			_, provider := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()

			if filePath == "" {
				showError(app, "File path is required")
				return
			}

			parseAndShowResults(app, filePath, provider)
		}).
		AddButton("Back", func() {
			showMainMenu(app)
		})

	form.SetBorder(true).
		SetTitle("Parse Expenses").
		SetTitleAlign(tview.AlignCenter)

	app.SetRoot(form, true)
}

func parseAndShowResults(app *tview.Application, filePath, provider string) {
	expenseCategories, parsedExpenses, err := parseAndAggregate(filePath, provider)
	if err != nil {
		showError(app, fmt.Sprintf("Error parsing expenses: %v", err))
		return
	}

	// Build result text
	var result strings.Builder
	result.WriteString("=== Grouped Categories ===\n\n")

	for _, category := range expenseCategories.categories {
		result.WriteString(fmt.Sprintf("%s: %.2f\n", category.Category, category.Amount))
		for _, expense := range category.Expenses {
			result.WriteString(fmt.Sprintf("  • %s: %.2f\n", expense.Description, expense.Amount))
		}
		result.WriteString("\n")
	}

	result.WriteString("\n=== Unmatched Expenses ===\n\n")
	for _, expense := range parsedExpenses {
		if !expense.Matched {
			result.WriteString(fmt.Sprintf("%s: %.2f\n", expense.Description, expense.Amount))
		}
	}

	fmt.Println(result.String())

	// Use the actual expense results with the working configuration
	contentToShow := result.String()
	if len(contentToShow) == 0 {
		contentToShow = "No expense data found or processed.\n\nThis could mean:\n- File is empty\n- Wrong provider selected\n- File format issue\n\nPress ESC to go back."
	}

	// Create textView with the working configuration
	textView := tview.NewTextView()
	textView.SetText(contentToShow)
	textView.SetScrollable(true) // Re-enable scrolling for real data
	textView.SetWrap(true)
	textView.SetBorder(true)
	textView.SetTitle("Expense Results")

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			showMainMenu(app)
		}
		return event
	})

	app.SetRoot(textView, true)
}

func showError(app *tview.Application, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			showMainMenu(app)
		})

	app.SetRoot(modal, true)
}

func showMainMenu(app *tview.Application) {
	menu := tview.NewList()
	menu.AddItem("List Supported Providers", "View all supported expense providers", 'l', func() {
		showSupportedProviders(app)
	})
	menu.AddItem("Parse Expenses", "Parse and categorize expenses from a file", 'p', func() {
		showParseExpensesForm(app)
	})
	menu.AddItem("Quit", "Exit the application", 'q', func() {
		app.Stop()
	})

	menu.SetBorder(true)
	menu.SetTitle("Expense Tracker")

	app.SetRoot(menu, true)
}

func parseExpenses(fileName string, provider string) ([]*Expense, error) {
	rawExpenses, err := readFile(fileName, provider)
	if err != nil {
		return nil, err
	}
	expenses := []*Expense{}

	var mapper func(map[string]string) (*Expense, error)

	if provider == "wise" {
		mapper = MapWiseExpense
	} else if provider == "revolut" {
		mapper = mapRevolutExpense
	} else if provider == "seb" {
		mapper = mapSEBExpense
	} else {
		return nil, fmt.Errorf("provider %s not supported", provider)
	}

	for _, row := range rawExpenses {
		expense, err := mapper(row)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func mapSlicesToMap(slices [][]string) ([]map[string]string, error) {
	csvMap := []map[string]string{}
	headers := slices[0]
	for _, slice := range slices[1:] {
		rowMap := make(map[string]string)
		for i, value := range slice {
			rowMap[headers[i]] = value
		}
		csvMap = append(csvMap, rowMap)
	}
	return csvMap, nil
}

func readFile(fileName string, provider string) ([]map[string]string, error) {
	absPath, _ := filepath.Abs(fileName)
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	if provider == "seb" {
		reader.LazyQuotes = true
		reader.Comma = ';'
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	csvMap, err := mapSlicesToMap(records)

	if err != nil {
		return nil, err
	}

	return csvMap, nil
}
