package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gregdel/pushover"
	"github.com/korovkin/limiter"
	_ "github.com/lib/pq"
)

const (
	connLimit = 800
	routines  = 400
)

func main() {
	connStr := os.Getenv("PG_CONN")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(connLimit)
	db.SetMaxIdleConns(600)
	db.SetConnMaxLifetime(time.Hour)

	// tarGzPath := "../tests/test_data.tar.gz"
	tarGzPath := os.Args[1]

	tarGz, err := os.Open(tarGzPath)
	if err != nil {
		log.Fatal(err)
	}
	defer tarGz.Close()

	gzf, err := gzip.NewReader(tarGz)
	if err != nil {
		log.Fatal(err)
	}

	tarReader := tar.NewReader(gzf)

	var counter int
	alert("Starting: " + tarGzPath)
	for {
		recordCountBefore := count(db, "records")
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		if header.Typeflag == tar.TypeReg {
			var wg sync.WaitGroup
			lineCh := make(chan string, routines)

			// process lines in the background as they come in to the lineCh channel
			// processing has not yet begun, but this 'listener' needs to be set up
			// first
			fmt.Println("Starting " + header.Name)
			limit := limiter.NewConcurrencyLimiter(routines)
			go func(
				wgi *sync.WaitGroup,
				dbi *sql.DB,
				ch chan string,
				lim *limiter.ConcurrencyLimiter,
			) {

				for {
					select {
					case line := <-ch:
						lim.Execute(func() {
							processAndSave(wgi, dbi, line)
						})
					}
				}

			}(&wg, db, lineCh, limit)

			// iterate through the lines in the file, adding each to the workgroup
			// before dispatching the line to the processing listener
			lineReader := bufio.NewScanner(tarReader)
			for lineReader.Scan() {
				lineCh <- lineReader.Text()
				counter++
				wg.Add(1)
			}

			fmt.Println("Waiting for " + header.Name)
			wg.Wait()
			recordCountAfter := count(db, "records")
			newRecordsCount := recordCountAfter - recordCountBefore
			duplicateCount := counter - newRecordsCount
			counter = 0

			msg := fmt.Sprintf(
				"Finished processing: %s\nNew: %d\nSkipped: %d\nTotal: %d",
				header.Name,
				newRecordsCount,
				duplicateCount,
				count(db, "records"),
			)
			alert(msg)
		}
	}
	alert("Finished: " + tarGzPath)
}

func processAndSave(wg *sync.WaitGroup, db *sql.DB, lineText string) {
	defer wg.Done()

	user, domain, password := parse(lineText)
	upsert(db, user, domain, password)
}

func upsert(db *sql.DB, user, domain, password string) {
	query := fmt.Sprintf(`
	WITH ins1 AS (
		INSERT INTO usernames(name) VALUES ('%s')
		ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name
		RETURNING id AS user_id
	)
	, ins2 AS (
		INSERT INTO passwords(password) VALUES ('%s')
		ON CONFLICT (password) DO UPDATE SET password=EXCLUDED.password
		RETURNING id AS pass_id
	)
	, ins3 AS (
		INSERT INTO domains(domain) VALUES ('%s')
		ON CONFLICT (domain) DO UPDATE SET domain=EXCLUDED.domain
		RETURNING id AS domain_id
	)

	INSERT INTO records (username_id, password_id, domain_id)
	VALUES (
		(select user_id from ins1), 
		(select pass_id from ins2), 
		(select domain_id from ins3) 
	)`, user, password, domain)

	tx, _ := db.Begin()
	_, err := tx.Exec(query)
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

func count(db *sql.DB, table string) int {
	var count int
	q := fmt.Sprintf(`SELECT COUNT(id) FROM %s`, table)
	rows, err := db.Query(q)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&count)
	}
	return count
}

func parse(line string) (user, domain, password string) {
	user, domain, password = "nil", "nil", "nil"

	userAndRest := strings.SplitN(line, "@", 2)
	if len(userAndRest) != 2 {
		return user, domain, password
	}

	if len(userAndRest) == 2 {
		user = userAndRest[0]
	}

	domainAndPass := strings.SplitN(userAndRest[1], ":", 2)
	if len(domainAndPass) == 2 {
		domain = domainAndPass[0]
		password = domainAndPass[1]
		return
	}

	domainAndPass = strings.SplitN(userAndRest[1], ";", 2)
	if len(domainAndPass) == 2 {
		domain = domainAndPass[0]
		password = domainAndPass[1]
		return
	}
	domain = domainAndPass[0]
	return
}

func alert(text string) {
	app := pushover.New(os.Getenv("PO_API"))
	me := pushover.NewRecipient(os.Getenv("PO_USR"))
	msg := pushover.NewMessage(text)
	app.SendMessage(msg, me)
}
