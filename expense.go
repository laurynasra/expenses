package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sfomuseum/go-csvdict/v2"
)

type Expense struct {
	Amount      float64
	Description string
	Date        time.Time
	Provider    string
	Category    string
}

type ExpenseCategory struct {
	Amount   float64
	Category string
	Matchers []string
}

type ExpenseCategories struct {
	categories []*ExpenseCategory
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

func parseAndAggregate(fileName string) error {
	expenseCategories := &ExpenseCategories{}
	expenseCategories.AddCategory(&ExpenseCategory{
		Amount:   0,
		Category: "Food",
		Matchers: []string{"maxima", "lidl"},
	})

	parsedExpenses, nil := parseExpenses(fileName)
	for _, expense := range parsedExpenses {
		for _, expenseCategory := range expenseCategories.categories {
			if expenseCategory.Match(strings.ToLower(expense.Description)) {
				expenseCategory.Amount += expense.Amount
				break //stop matching further categories
			}
		}
	}
	fmt.Println(expenseCategories)
	return nil
}

func main() {
	parseAndAggregate("wise-05.csv")
	// supportedProviders := []string{"Wise", "SEB", "Revolut"}
	// cmd := &cli.Command{
	// 	Commands: []*cli.Command{
	// 		{
	// 			Name:  "list-supported",
	// 			Usage: "Lists supported providers",
	// 			Action: func(ctx context.Context, cmd *cli.Command) error {
	// 				fmt.Println("Supported providers:", supportedProviders)
	// 				return nil
	// 			},
	// 		},
	// 		{
	// 			Name:  "parse-expenses",
	// 			Usage: "Parses expense report for given provider. Prints out aggregated and categorized expenses",
	// 			Flags: []cli.Flag{
	// 				&cli.StringFlag{
	// 					Name:     "provider",
	// 					Usage:    "Provider to parse expenses for",
	// 					Required: true,
	// 				},
	// 				&cli.StringFlag{
	// 					Name:     "file",
	// 					Usage:    "File to parse expenses from",
	// 					Required: true,
	// 				},
	// 			},
	// 			Action: func(ctx context.Context, cmd *cli.Command) error {
	// 				fileName := cmd.String("file")
	// 				return parseAndAggregate(fileName)
	// 			},
	// 		},
	// 	},
	// }
	// if err := cmd.Run(context.Background(), os.Args); err != nil {
	// 	log.Fatal(err)
	// }
}

func parseExpenses(fileName string) ([]*Expense, error) {
	// _ := cmd.String("provider")
	// fileName := cmd.String("file")

	r, err := readFile(fileName)
	if err != nil {
		return nil, err
	}

	expenses := []*Expense{}

	for row, err := range r.Iterate() {
		if err != nil {
			return nil, err
		}
		// fmt.Println("Amount:", row["Amount"], "Description:", row["Description"])
		expense, err := MapWiseExpense(row)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func readFile(fileName string) (*csvdict.Reader, error) {
	r, err := csvdict.NewReaderFromPath(fileName)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func AggregateExpenses(fileName string) error {
	expenseCategories := &ExpenseCategories{}
	expenseCategories.AddCategory(&ExpenseCategory{
		Amount:   0,
		Category: "Food",
		Matchers: []string{"maxima", "lidl"},
	})

	expenses, err := parseExpensesDirectly(fileName)
	if err != nil {
		return err
	}

	for _, expense := range expenses {
		for _, expenseCategory := range expenseCategories.categories {
			if expenseCategory.Match(expense.Description) {
				expenseCategory.Amount += expense.Amount
				break //stop matching further categories
			}
		}
	}
	fmt.Println(expenseCategories)
	return nil
}

func parseExpensesDirectly(fileName string) ([]*Expense, error) {
	r, err := readFile(fileName)
	if err != nil {
		return nil, err
	}

	expenses := []*Expense{}

	for row, err := range r.Iterate() {
		if err != nil {
			return nil, err
		}
		expense, err := MapWiseExpense(row)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}
