package api

import (
	"log"
	"net/http"
)

func StartServer(port, categoriesPath string) error {
	store := NewSessionStore()
	h := &Handler{
		store:          store,
		categoriesPath: categoriesPath,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/upload", h.Upload)
	mux.HandleFunc("GET /api/sessions/{id}", h.GetSession)
	mux.HandleFunc("GET /api/categories", h.ListCategories)
	mux.HandleFunc("POST /api/categories/{name}/matchers", h.AddMatcher)
	mux.HandleFunc("POST /api/expenses/{sessionId}/categorize", h.CategorizeExpense)

	log.Printf("API server listening on :%s", port)
	return http.ListenAndServe(":"+port, corsMiddleware(mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
