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

const connLimit = 450

func main() {
	connStr := os.Getenv("PG_CONN")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(connLimit)
	db.SetMaxIdleConns(200)
	db.SetConnMaxLifetime(time.Hour)

	// tarGzPath := "test.tar.gz"
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
			lineCh := make(chan string, connLimit)

			// process lines in the background as they come in to the lineCh channel
			// processing has not yet begun, but this 'listener' needs to be set up
			// first
			fmt.Println("Starting " + header.Name)
			limit := limiter.NewConcurrencyLimiter(connLimit)
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
	userID, userExists := findOrCreate(db, "usernames", "name", strings.ToLower(user))
	domainID, domainExists := findOrCreate(db, "domains", "domain", strings.ToLower(domain))
	passwordID, passwordExists := findOrCreate(db, "passwords", "password", password)

	if userExists && domainExists && passwordExists {
		createJoin(db, userID, passwordID, domainID)
	}
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

func findOrCreate(db *sql.DB, table, column, attr string) (int, bool) {
	var id int

	q := fmt.Sprintf(`SELECT id from %s WHERE %s = '%s'`, table, column, attr)
	statement, err := db.Prepare(q)
	if err != nil {
		return id, false
	}
	defer statement.Close()

	rows, err := statement.Query()
	if err != nil {
		return id, false
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return id, false
		}
		return id, true
	}

	newID, err := create(db, table, column, attr)
	if err != nil {
		return newID, false
	}
	return newID, true
}

func create(db *sql.DB, table, col, attr string) (int, error) {
	var id int
	q := fmt.Sprintf(
		`INSERT INTO %s(%s) VALUES ('%s') RETURNING id`,
		table,
		col,
		attr,
	)
	err := db.QueryRow(q).Scan(&id)
	return id, err
}

func createJoin(db *sql.DB, userID, passID, domainID int) (int, error) {
	q := fmt.Sprintf(
		`INSERT INTO records(username_id, password_id, domain_id) 
		VALUES ('%d', '%d', '%d') RETURNING id`,
		userID,
		passID,
		domainID,
	)

	var id int
	err := db.QueryRow(q).Scan(&id)
	return id, err
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
