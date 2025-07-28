package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

var allowedOrigins = map[string]bool{
	"http://localhost:1313":             true,
	"http://localhost:8080":             true,
	"https://eaglesoneats.com":          true,
	"https://commento.eaglesoneats.com": true,
}

func recipeRatingHandler(w http.ResponseWriter, r *http.Request) {
	type RatingRequest struct {
		Domain string `json:"domain"`
		Path   string `json:"path"`
	}

	origin := r.Header.Get("Origin")

	if allowedOrigins[origin] {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RatingRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": "error parsing json request - " + err.Error()})
		return
	}

	if req.Domain == "" || req.Path == "" {
		bodyMarshal(w, response{"success": false, "message": "missing domain or path"})
		return
	}

	domain := domainStrip(req.Domain)

	d, err := domainGet(domain)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	if d.State == "frozen" {
		bodyMarshal(w, response{"success": false, "message": errorDomainFrozen.Error()})
		return
	}

	var total int
	var average sql.NullFloat64

	sqlQuery := `
SELECT COUNT(rating) AS total, AVG(rating) AS average
FROM comments
WHERE state= 'approved' AND domain = $1 AND path = $2 AND rating IS NOT NULL;
	`

	err = db.QueryRow(sqlQuery, domain, req.Path).Scan(&total, &average)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": "query error - " + err.Error()})
		return
	}

	avg := 0.0
	if average.Valid {
		avg = average.Float64
	}

	bodyMarshal(w, response{"success": true, "rating": avg, "total": total})
}
