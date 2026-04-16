package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExpenseCategory_Match(t *testing.T) {
	tests := []struct {
		name        string
		matchers    []string
		description string
		want        bool
	}{
		{
			name:        "exact match lowercase",
			matchers:    []string{"groceries", "food"},
			description: "groceries at store",
			want:        true,
		},
		{
			name:        "case insensitive match",
			matchers:    []string{"Groceries"},
			description: "groceries at store",
			want:        true,
		},
		{
			name:        "partial match",
			matchers:    []string{"uber"},
			description: "Uber trip to downtown",
			want:        true,
		},
		{
			name:        "no match",
			matchers:    []string{"groceries", "food"},
			description: "gas station",
			want:        false,
		},
		{
			name:        "empty matchers",
			matchers:    []string{},
			description: "anything",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ExpenseCategory{Matchers: tt.matchers}
			if got := e.Match(tt.description); got != tt.want {
				t.Errorf("ExpenseCategory.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapWiseExpense(t *testing.T) {
	tests := []struct {
		name    string
		row     map[string]string
		want    *Expense
		wantErr bool
	}{
		{
			name: "valid expense",
			row:  map[string]string{"Amount": "50.00", "Description": "Coffee shop"},
			want: &Expense{Amount: -50.00, Description: "Coffee shop", Provider: "Wise"},
		},
		{
			name:    "invalid amount",
			row:     map[string]string{"Amount": "invalid", "Description": "Coffee shop"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapWiseExpense(tt.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapWiseExpense() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Amount != tt.want.Amount {
					t.Errorf("MapWiseExpense() Amount = %v, want %v", got.Amount, tt.want.Amount)
				}
				if got.Description != tt.want.Description {
					t.Errorf("MapWiseExpense() Description = %v, want %v", got.Description, tt.want.Description)
				}
				if got.Provider != tt.want.Provider {
					t.Errorf("MapWiseExpense() Provider = %v, want %v", got.Provider, tt.want.Provider)
				}
			}
		})
	}
}

func TestMapRevolutExpense(t *testing.T) {
	tests := []struct {
		name    string
		row     map[string]string
		want    *Expense
		wantErr bool
	}{
		{
			name: "valid expense",
			row:  map[string]string{"Amount": "25.50", "Description": "Restaurant"},
			want: &Expense{Amount: 25.50, Description: "Restaurant", Provider: "Revolut"},
		},
		{
			name: "negative amount",
			row:  map[string]string{"Amount": "-10.00", "Description": "Refund"},
			want: &Expense{Amount: -10.00, Description: "Refund", Provider: "Revolut"},
		},
		{
			name:    "invalid amount",
			row:     map[string]string{"Amount": "abc", "Description": "Restaurant"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapRevolutExpense(tt.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapRevolutExpense() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Amount != tt.want.Amount {
					t.Errorf("MapRevolutExpense() Amount = %v, want %v", got.Amount, tt.want.Amount)
				}
				if got.Description != tt.want.Description {
					t.Errorf("MapRevolutExpense() Description = %v, want %v", got.Description, tt.want.Description)
				}
				if got.Provider != tt.want.Provider {
					t.Errorf("MapRevolutExpense() Provider = %v, want %v", got.Provider, tt.want.Provider)
				}
			}
		})
	}
}

func TestMapSEBExpense(t *testing.T) {
	tests := []struct {
		name    string
		row     map[string]string
		want    *Expense
		wantErr bool
	}{
		{
			name: "valid expense with comma decimal",
			row: map[string]string{
				"SUMA":                             "100,50",
				"MOKĖTOJO ARBA GAVĖJO PAVADINIMAS": "Store Name",
				"MOKĖJIMO PASKIRTIS":               "Purchase",
				"TRANSAKCIJOS TIPAS":               "Debit",
			},
			want: &Expense{Amount: 100.50, Description: "Store Name Purchase Debit", Provider: "SEB"},
		},
		{
			name: "valid expense with dot decimal",
			row: map[string]string{
				"SUMA":                             "50.25",
				"MOKĖTOJO ARBA GAVĖJO PAVADINIMAS": "Vendor",
				"MOKĖJIMO PASKIRTIS":               "Service",
				"TRANSAKCIJOS TIPAS":               "Credit",
			},
			want: &Expense{Amount: 50.25, Description: "Vendor Service Credit", Provider: "SEB"},
		},
		{
			name: "invalid amount",
			row: map[string]string{
				"SUMA":                             "invalid",
				"MOKĖTOJO ARBA GAVĖJO PAVADINIMAS": "Store",
				"MOKĖJIMO PASKIRTIS":               "Purchase",
				"TRANSAKCIJOS TIPAS":               "Debit",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapSEBExpense(tt.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapSEBExpense() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Amount != tt.want.Amount {
					t.Errorf("MapSEBExpense() Amount = %v, want %v", got.Amount, tt.want.Amount)
				}
				if got.Description != tt.want.Description {
					t.Errorf("MapSEBExpense() Description = %v, want %v", got.Description, tt.want.Description)
				}
				if got.Provider != tt.want.Provider {
					t.Errorf("MapSEBExpense() Provider = %v, want %v", got.Provider, tt.want.Provider)
				}
			}
		})
	}
}

func TestMapSlicesToMap(t *testing.T) {
	tests := []struct {
		name    string
		slices  [][]string
		want    []map[string]string
		wantErr bool
	}{
		{
			name: "valid CSV data",
			slices: [][]string{
				{"Name", "Amount", "Date"},
				{"Expense1", "100", "2024-01-01"},
				{"Expense2", "200", "2024-01-02"},
			},
			want: []map[string]string{
				{"Name": "Expense1", "Amount": "100", "Date": "2024-01-01"},
				{"Name": "Expense2", "Amount": "200", "Date": "2024-01-02"},
			},
		},
		{
			name:    "empty data with headers only",
			slices:  [][]string{{"Name", "Amount"}},
			want:    []map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapSlicesToMap(tt.slices)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapSlicesToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("MapSlicesToMap() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				for key := range tt.want[i] {
					if got[i][key] != tt.want[i][key] {
						t.Errorf("MapSlicesToMap() row %d, key %s = %v, want %v", i, key, got[i][key], tt.want[i][key])
					}
				}
			}
		})
	}
}

func TestExpenseCategories_AddCategory(t *testing.T) {
	ec := &ExpenseCategories{}
	category := &ExpenseCategory{
		Category: "Food",
		Amount:   0,
		Matchers: []string{"restaurant", "groceries"},
	}

	ec.AddCategory(category)

	if len(ec.Categories) != 1 {
		t.Errorf("AddCategory() categories length = %v, want 1", len(ec.Categories))
	}
	if ec.Categories[0].Category != "Food" {
		t.Errorf("AddCategory() category name = %v, want Food", ec.Categories[0].Category)
	}
}

func TestLoadCategoriesFromJSON(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_categories.json")

	categories := []CategoryConfig{
		{Name: "Food", Matchers: []string{"restaurant", "groceries"}},
		{Name: "Transport", Matchers: []string{"uber", "taxi"}},
	}

	data, err := json.Marshal(categories)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	if err := os.WriteFile(testFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	ec, err := LoadCategoriesFromJSON(testFile)
	if err != nil {
		t.Errorf("LoadCategoriesFromJSON() error = %v", err)
		return
	}

	if len(ec.Categories) != 2 {
		t.Errorf("LoadCategoriesFromJSON() categories length = %v, want 2", len(ec.Categories))
	}
	if ec.Categories[0].Category != "Food" {
		t.Errorf("LoadCategoriesFromJSON() first category = %v, want Food", ec.Categories[0].Category)
	}
	if len(ec.Categories[0].Matchers) != 2 {
		t.Errorf("LoadCategoriesFromJSON() first category matchers length = %v, want 2", len(ec.Categories[0].Matchers))
	}
}

func TestLoadCategoriesFromJSON_FileNotFound(t *testing.T) {
	_, err := LoadCategoriesFromJSON("nonexistent.json")
	if err == nil {
		t.Error("LoadCategoriesFromJSON() expected error for nonexistent file, got nil")
	}
}

func TestLoadCategoriesFromJSON_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(testFile, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadCategoriesFromJSON(testFile)
	if err == nil {
		t.Error("LoadCategoriesFromJSON() expected error for invalid JSON, got nil")
	}
}

func TestParseExpenses_UnsupportedProvider(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")
	if err := os.WriteFile(testFile, []byte("Amount,Description\n100,Test\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := ParseExpenses(testFile, "unsupported")
	if err == nil {
		t.Error("ParseExpenses() expected error for unsupported provider, got nil")
	}
}
