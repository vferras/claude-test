package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"claude-test/analysis"
)

func NewServer(db *sql.DB, port string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /analysis", func(w http.ResponseWriter, r *http.Request) {
		result, err := analysis.Analyze(db)
		if err != nil {
			log.Printf("Analysis error: %v", err)
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	return &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
}
