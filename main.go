package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	hibp "github.com/audibleblink/haveibeenpwned"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/bigquery"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	projectID     = os.Getenv("GOOGLE_CLOUD_PROJECT")
	bigQueryTable = os.Getenv("GOOGLE_BIGQUERY_TABLE")
	googleCred    = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	hibpKey       = os.Getenv("HIBP_API_KEY")

	bq *bigquery.Client
)

func init() {
	var err error
	if projectID == "" || bigQueryTable == "" || googleCred == "" || hibpKey == "" {
		err = fmt.Errorf("missing required environment variables")
		log.Fatal(err)
	}

	ctx := context.Background()
	bq, err = bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/usernames/{username}", handleUsername)
	r.Get("/passwords/{password}", handlePassword)
	r.Get("/domains/{domain}", handleDomain)
	r.Get("/emails/{email}", handleEmail)
	r.Get("/breaches/{email}", handleBreaches)

	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}

type Record struct {
	Username string
	Domain   string
	Password string
}

type Breach struct {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resultWriter(w, &records)
}

func handlePassword(w http.ResponseWriter, r *http.Request) {
	password := chi.URLParam(r, "password")
	records, err := recordsByPassword(password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resultWriter(w, &records)
}

func handleDomain(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	records, err := recordsByDomain(domain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resultWriter(w, &records)
}

func handleEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	records, err := recordsByEmail(email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resultWriter(w, &records)
}

func handleBreaches(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	hibpBreaches, err := hibp.BreachedAccount(email, "", false, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var breaches []*Breach
	for _, hibpBreach := range hibpBreaches {
		breach := &Breach{
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

func recordsByUsername(username string) (records []*Record, err error) {
	return recordsBy("username", username)
}

func recordsByPassword(password string) (records []*Record, err error) {
	return recordsBy("password", password)
}

func recordsByDomain(domain string) (records []*Record, err error) {
	return recordsBy("domain", domain)
}

func recordsByEmail(email string) (records []*Record, err error) {
	usernameAndDomain := strings.Split(email, "@")
	if len(usernameAndDomain) != 2 {
		err = fmt.Errorf("invalid email format")
		return
	}

	username, domain := usernameAndDomain[0], usernameAndDomain[1]
	queryString := fmt.Sprintf(`SELECT DISTINCT * FROM %s WHERE username = "%s" AND domain = "%s"`, bigQueryTable, username, domain)
	return queryRecords(queryString)
}

func queryRecords(queryString string) (records []*Record, err error) {
	query := bq.Query(queryString)
	ctx := context.Background()
	results, err := query.Read(ctx)
	if err != nil {
		return
	}

	for {
		var r Record
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

func recordsBy(column, value string) (records []*Record, err error) {
	queryString := fmt.Sprintf(`SELECT DISTINCT * FROM %s WHERE %s = "%s"`, bigQueryTable, column, value)
	return queryRecords(queryString)
}

func resultWriter(w http.ResponseWriter, records *[]*Record) {
	resultJson, err := json.Marshal(records)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resultJson)
}