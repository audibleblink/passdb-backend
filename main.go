package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"

	"github.com/audibleblink/passdb/hibp"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

var (
	projectID     = os.Getenv("GOOGLE_CLOUD_PROJECT")
	bigQueryTable = os.Getenv("GOOGLE_BIGQUERY_TABLE")
	googleCred    = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	hibpKey       = os.Getenv("HIBP_API_KEY")

	listenAddr = ":3000"
	bq         *bigquery.Client
)

func init() {
	var err error
	if projectID == "" || bigQueryTable == "" || googleCred == "" || hibpKey == "" {
		err = fmt.Errorf("missing required environment variables")
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		listenAddr = os.Args[1]
	}

	ctx := context.Background()
	bq, err = bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	cacheConfig := LoadCacheConfig()

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(CacheMiddleware(cacheConfig))
	r.Use(middleware.Recoverer)

	// API routes with versioning
	r.Route("/api/v1", func(r chi.Router) {
		// Health check endpoint
		r.Get("/health", handleHealth)

		// Password database endpoints
		r.Get("/usernames/{username}", handleUsername)
		r.Get("/passwords/{password}", handlePassword)
		r.Get("/domains/{domain}", handleDomain)
		r.Get("/emails/{email}", handleEmail)
		r.Get("/breaches/{email}", handleBreaches)

		// Cache management endpoints
		r.Get("/cache/stats", handleCacheStats)
		r.Delete("/cache", handleCacheClear)
		r.Delete("/cache/{pattern}", handleCacheClearPattern)
	})

	log.Printf("Starting server on %s\n", listenAddr)
	log.Printf("API endpoints available at /api/v1/")
	log.Printf("Static files served from /")
	err := http.ListenAndServe(listenAddr, r)
	if err != nil {
		log.Fatal(err)
	}
}

type record struct {
	Username bigquery.NullString `json:"username"`
	Domain   bigquery.NullString `json:"domain"`
	Password bigquery.NullString `json:"password"`
}

type breach struct {
	Title       string
	Domain      string
	Date        string
	Count       int
	Description string
	LogoPath    string
}

func handleUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	records, err := recordsByUsername(username)
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}
	resultWriter(w, records)
}

func handlePassword(w http.ResponseWriter, r *http.Request) {
	password := chi.URLParam(r, "password")
	records, err := recordsByPassword(password)
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}
	resultWriter(w, records)
}

func handleDomain(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	records, err := recordsByDomain(domain)
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}
	resultWriter(w, records)
}

func handleEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	records, err := recordsByEmail(email)
	if err != nil {
		JSONError(w, err, http.StatusBadRequest)
		return
	}
	resultWriter(w, records)
}

func handleBreaches(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	hibpBreaches, err := hibp.BreachedAccount(email, "", false, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var breaches []*breach
	for _, hibpBreach := range hibpBreaches {
		breach := &breach{
			Title:       hibpBreach.Title,
			Domain:      hibpBreach.Domain,
			Date:        hibpBreach.BreachDate,
			Count:       hibpBreach.PwnCount,
			Description: hibpBreach.Description,
			LogoPath:    hibpBreach.LogoPath,
		}
		breaches = append(breaches, breach)
	}

	data, err := json.Marshal(breaches)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func recordsByUsername(username string) (records []*record, err error) {
	return recordsBy("username", username)
}

func recordsByPassword(password string) (records []*record, err error) {
	return recordsBy("password", password)
}

func recordsByDomain(domain string) (records []*record, err error) {
	return recordsBy("domain", domain)
}

func recordsByEmail(email string) (records []*record, err error) {
	usernameAndDomain := strings.Split(email, "@")
	if len(usernameAndDomain) != 2 {
		err = fmt.Errorf("invalid email format")
		return
	}

	queryString := fmt.Sprintf(
		`SELECT DISTINCT * FROM %s WHERE username = @username AND domain = @domain`,
		bigQueryTable,
	)

	params := map[string]string{
		"username": usernameAndDomain[0],
		"domain":   usernameAndDomain[1],
	}
	query := parameterize(queryString, params)
	return queryRecords(query)
}

func recordsBy(column, value string) (records []*record, err error) {
	queryString := fmt.Sprintf(
		`SELECT DISTINCT * FROM %s WHERE %s = @%s`,
		bigQueryTable,
		column,
		column,
	)
	params := map[string]string{column: value}
	query := parameterize(queryString, params)
	return queryRecords(query)
}

func parameterize(q string, fields map[string]string) *bigquery.Query {
	var params []bigquery.QueryParameter
	for key, value := range fields {
		param := bigquery.QueryParameter{Name: key, Value: value}
		params = append(params, param)
	}
	query := bq.Query(q)
	query.Parameters = params
	return query
}

func queryRecords(query *bigquery.Query) (records []*record, err error) {
	records = make([]*record, 0)
	ctx := context.Background()
	results, err := query.Read(ctx)
	if err != nil {
		return
	}

	for {
		var r record
		err = results.Next(&r)
		if err == iterator.Done {
			err = nil
			break
		}
		if err != nil {
			return
		}
		records = append(records, &r)
	}
	return
}

func resultWriter(w http.ResponseWriter, records []*record) {
	resultJSON, err := json.Marshal(records)
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	w.Write(resultJSON)
}

type JSONErr struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func JSONError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	error := &JSONErr{code, err.Error()}
	json.NewEncoder(w).Encode(error)
}

func handleCacheStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats, err := GetCacheStats()
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func handleCacheClear(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := ClearCache(); err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"message": "Cache cleared successfully",
		"success": true,
	}

	json.NewEncoder(w).Encode(response)
}

func handleCacheClearPattern(w http.ResponseWriter, r *http.Request) {
	pattern := chi.URLParam(r, "pattern")
	w.Header().Set("Content-Type", "application/json")

	deletedCount, err := ClearCachePattern(pattern)
	if err != nil {
		JSONError(w, err, http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"message":       fmt.Sprintf("Cache entries matching pattern '%s' cleared", pattern),
		"deleted_count": deletedCount,
		"success":       true,
	}

	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]any{
		"status":  "healthy",
		"service": "passdb-api",
		"version": "v1",
	}

	json.NewEncoder(w).Encode(response)
}
