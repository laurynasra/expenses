package api

type ExpenseDTO struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Provider    string  `json:"provider"`
	Category    string  `json:"category"`
	Matched     bool    `json:"matched"`
}

type CategoryDTO struct {
	Name     string       `json:"name"`
	Amount   float64      `json:"amount"`
	Matchers []string     `json:"matchers"`
	Expenses []ExpenseDTO `json:"expenses"`
}

type SessionDTO struct {
	ID                    string        `json:"id"`
	Provider              string        `json:"provider"`
	Categories            []CategoryDTO `json:"categories"`
	UncategorizedExpenses []ExpenseDTO  `json:"uncategorized"`
	TotalAmount           float64       `json:"totalAmount"`
}

type AddMatcherRequest struct {
	Matcher string `json:"matcher"`
}

type CategorizeExpenseRequest struct {
	ExpenseID    string `json:"expenseId"`
	CategoryName string `json:"categoryName"`
	Matcher      string `json:"matcher"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
