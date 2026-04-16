package engine

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	Categories []*ExpenseCategory
}

type CategoryConfig struct {
	Name     string   `json:"name"`
	Matchers []string `json:"matchers"`
}

func LoadCategoriesFromJSON(filename string) (*ExpenseCategories, error) {
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

func SaveCategories(filename string, expenseCategories *ExpenseCategories) error {
	configs := make([]CategoryConfig, 0, len(expenseCategories.Categories))
	for _, cat := range expenseCategories.Categories {
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

func (e *ExpenseCategories) AddCategory(expense *ExpenseCategory) {
	e.Categories = append(e.Categories, expense)
}

func (e *ExpenseCategory) Match(description string) bool {
	for _, matcher := range e.Matchers {
		if strings.Contains(strings.ToLower(description), strings.ToLower(matcher)) {
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

	return &Expense{
		Amount:      amount,
		Description: row["Description"],
		Provider:    "Wise",
	}, nil
}

func MapRevolutExpense(row map[string]string) (*Expense, error) {
	amount, err := strconv.ParseFloat(row["Amount"], 64)
	if err != nil {
		return nil, err
	}
	return &Expense{
		Amount:      amount,
		Description: row["Description"],
		Provider:    "Revolut",
	}, nil
}

func MapSEBExpense(row map[string]string) (*Expense, error) {
	amount, err := strconv.ParseFloat(strings.ReplaceAll(row["SUMA"], ",", "."), 64)
	if err != nil {
		return nil, err
	}
	description := strings.Join([]string{
		row["MOKĖTOJO ARBA GAVĖJO PAVADINIMAS"],
		row["MOKĖJIMO PASKIRTIS"],
		row["TRANSAKCIJOS TIPAS"],
	}, " ")

	return &Expense{
		Amount:      amount,
		Description: description,
		Provider:    "SEB",
	}, nil
}

func ParseAndAggregate(fileName, provider, categoriesPath string) (*ExpenseCategories, []*Expense, error) {
	expenseCategories, err := LoadCategoriesFromJSON(categoriesPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load categories: %w", err)
	}

	parsedExpenses, err := ParseExpenses(fileName, provider)
	if err != nil {
		return nil, nil, err
	}

	for _, expense := range parsedExpenses {
		for _, expenseCategory := range expenseCategories.Categories {
			if expenseCategory.Match(strings.ToLower(expense.Description)) {
				expenseCategory.Amount += expense.Amount
				expense.Matched = true
				expenseCategory.Expenses = append(expenseCategory.Expenses, expense)
				break
			}
		}
	}

	return expenseCategories, parsedExpenses, nil
}

func ParseExpenses(fileName, provider string) ([]*Expense, error) {
	rawExpenses, err := ReadFile(fileName, provider)
	if err != nil {
		return nil, err
	}

	var mapper func(map[string]string) (*Expense, error)
	switch provider {
	case "wise":
		mapper = MapWiseExpense
	case "revolut":
		mapper = MapRevolutExpense
	case "seb":
		mapper = MapSEBExpense
	default:
		return nil, fmt.Errorf("provider %s not supported", provider)
	}

	expenses := make([]*Expense, 0, len(rawExpenses))
	for _, row := range rawExpenses {
		expense, err := mapper(row)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func MapSlicesToMap(slices [][]string) ([]map[string]string, error) {
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

func ReadFile(fileName, provider string) ([]map[string]string, error) {
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

	return MapSlicesToMap(records)
}
