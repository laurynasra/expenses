package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"laurynasra/expenses/internal/engine"
)

type Handler struct {
	store          *SessionStore
	categoriesPath string
	catMu          sync.Mutex
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	provider := r.FormValue("provider")
	if provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	tmp, err := os.CreateTemp("", "expense-*.csv")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create temp file")
		return
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.ReadFrom(file); err != nil {
		tmp.Close()
		writeError(w, http.StatusInternalServerError, "failed to write temp file")
		return
	}
	tmp.Close()

	cats, allExpenses, err := engine.ParseAndAggregate(tmp.Name(), provider, h.categoriesPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse: %v", err))
		return
	}

	session := buildSession(provider, cats, allExpenses)
	h.store.Create(session)
	writeJSON(w, http.StatusOK, session)
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	session, ok := h.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := engine.LoadCategoriesFromJSON(h.categoriesPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load categories: %v", err))
		return
	}
	dtos := make([]CategoryDTO, 0, len(cats.Categories))
	for _, cat := range cats.Categories {
		matchers := cat.Matchers
		if matchers == nil {
			matchers = make([]string, 0)
		}
		dtos = append(dtos, CategoryDTO{
			Name:     cat.Category,
			Amount:   cat.Amount,
			Matchers: matchers,
			Expenses: make([]ExpenseDTO, 0),
		})
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *Handler) AddMatcher(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	var req AddMatcherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Matcher == "" {
		writeError(w, http.StatusBadRequest, "matcher is required")
		return
	}

	h.catMu.Lock()
	defer h.catMu.Unlock()

	cats, err := engine.LoadCategoriesFromJSON(h.categoriesPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load categories")
		return
	}

	var found *engine.ExpenseCategory
	for _, cat := range cats.Categories {
		if strings.EqualFold(cat.Category, name) {
			found = cat
			break
		}
	}
	if found == nil {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	found.Matchers = append(found.Matchers, req.Matcher)
	if err := engine.SaveCategories(h.categoriesPath, cats); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save categories")
		return
	}

	writeJSON(w, http.StatusOK, CategoryDTO{
		Name:     found.Category,
		Amount:   found.Amount,
		Matchers: found.Matchers,
	})
}

func (h *Handler) CategorizeExpense(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionId")
	session, ok := h.store.Get(sessionID)
	if !ok {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	var req CategorizeExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	expIdx := -1
	for i, exp := range session.UncategorizedExpenses {
		if exp.ID == req.ExpenseID {
			expIdx = i
			break
		}
	}
	if expIdx == -1 {
		writeError(w, http.StatusNotFound, "expense not found")
		return
	}

	catIdx := -1
	for i, cat := range session.Categories {
		if strings.EqualFold(cat.Name, req.CategoryName) {
			catIdx = i
			break
		}
	}
	if catIdx == -1 {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	exp := session.UncategorizedExpenses[expIdx]
	exp.Category = req.CategoryName
	exp.Matched = true
	session.UncategorizedExpenses = append(
		session.UncategorizedExpenses[:expIdx],
		session.UncategorizedExpenses[expIdx+1:]...,
	)
	session.Categories[catIdx].Expenses = append(session.Categories[catIdx].Expenses, exp)
	session.Categories[catIdx].Amount += exp.Amount

	if req.Matcher != "" {
		h.catMu.Lock()
		cats, err := engine.LoadCategoriesFromJSON(h.categoriesPath)
		if err == nil {
			for _, cat := range cats.Categories {
				if strings.EqualFold(cat.Category, req.CategoryName) {
					cat.Matchers = append(cat.Matchers, req.Matcher)
					if err := engine.SaveCategories(h.categoriesPath, cats); err == nil {
						session.Categories[catIdx].Matchers = cat.Matchers
					}
					break
				}
			}
		}
		h.catMu.Unlock()
	}

	h.store.Update(sessionID, session)
	writeJSON(w, http.StatusOK, session)
}

func buildSession(provider string, cats *engine.ExpenseCategories, allExpenses []*engine.Expense) *SessionDTO {
	session := &SessionDTO{
		Provider:              provider,
		Categories:            make([]CategoryDTO, 0),
		UncategorizedExpenses: make([]ExpenseDTO, 0),
	}
	globalIdx := 0

	for _, cat := range cats.Categories {
		catDTO := CategoryDTO{
			Name:     cat.Category,
			Amount:   cat.Amount,
			Matchers: cat.Matchers,
			Expenses: make([]ExpenseDTO, 0),
		}
		for _, exp := range cat.Expenses {
			catDTO.Expenses = append(catDTO.Expenses, toExpenseDTO(exp, fmt.Sprintf("%d", globalIdx)))
			globalIdx++
		}
		session.Categories = append(session.Categories, catDTO)
		session.TotalAmount += cat.Amount
	}

	for _, exp := range allExpenses {
		if !exp.Matched {
			session.UncategorizedExpenses = append(session.UncategorizedExpenses,
				toExpenseDTO(exp, fmt.Sprintf("%d", globalIdx)))
			globalIdx++
		}
	}

	return session
}

func toExpenseDTO(e *engine.Expense, id string) ExpenseDTO {
	return ExpenseDTO{
		ID:          id,
		Amount:      e.Amount,
		Description: e.Description,
		Date:        e.Date.Format(time.RFC3339),
		Provider:    e.Provider,
		Category:    e.Category,
		Matched:     e.Matched,
	}
}
