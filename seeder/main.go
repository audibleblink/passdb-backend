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
	connLimit = 100
	routines  = 50
)

var (
	doneLog  = "done.log"
	errLog   = "done.err"
	finished map[string]bool
)

func init() {
	progressFile, err := os.Open(doneLog)
	finished = make(map[string]bool)
	if err != nil {
		panic(err)
	}
	defer progressFile.Close()

	fileScanner := bufio.NewScanner(progressFile)
	for fileScanner.Scan() {
		f := fileScanner.Text()
		finished[f] = true
	}
}

func main() {
	f, err := os.OpenFile(errLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	connStr := os.Getenv("PG_CONN")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(connLimit)
	db.SetMaxIdleConns(connLimit)
	db.SetConnMaxLifetime(connLimit * time.Second)

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
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		if header.Typeflag == tar.TypeReg {
			if alreadyRun(header.Name) {
				fmt.Printf("Skipping: %s\n", header.Name)
				continue
			}

			var wg sync.WaitGroup
			lineCh := make(chan string, routines)

			// process lines in the background as they come in to the lineCh channel
			// processing has not yet begun, but this 'listener' needs to be set up
			// first
			fmt.Println("Starting " + header.Name)
			recordCountBefore := count(db, "records")
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
			counter = 0
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

			msg := fmt.Sprintf(
				"Finished processing: %s\nNew: %d\nSkipped: %d\nTotal: %d",
				header.Name,
				newRecordsCount,
				duplicateCount,
				recordCountAfter,
			)
			alert(msg)
			markDone(header.Name)
		}
	}
	alert("Completed tar: " + tarGzPath)
}

func processAndSave(wg *sync.WaitGroup, db *sql.DB, lineText string) {
	defer wg.Done()

	user, domain, password := parse(lineText)
	err := upsert(db, user, domain, password)

	if err != nil {
		log.Printf("COMMIT %s - %s", lineText, err.Error())
	}
}

func upsert(db *sql.DB, user, domain, password string) error {
	query := `
	WITH ins1 AS (
		INSERT INTO usernames(name) VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name
		RETURNING id AS user_id
	)
	, ins2 AS (
		INSERT INTO passwords(password) VALUES ($2)
		ON CONFLICT (password) DO UPDATE SET password=EXCLUDED.password
		RETURNING id AS pass_id
	)
	, ins3 AS (
		INSERT INTO domains(domain) VALUES ($3)
		ON CONFLICT (domain) DO UPDATE SET domain=EXCLUDED.domain
		RETURNING id AS domain_id
	)

	INSERT INTO records (username_id, password_id, domain_id)
	VALUES (
		(select user_id from ins1), 
		(select pass_id from ins2), 
		(select domain_id from ins3) 
	)`

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	preparedQuery, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer preparedQuery.Close()

	_, err = preparedQuery.Exec(user, password, domain)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// not the best way to count, but select COUNT(id) takes
// an increaseing amount of time as the # of records grows
func count(db *sql.DB, table string) int {
	var id int
	q := fmt.Sprintf(`SELECT MAX(id) FROM %s`, table)
	rows, err := db.Query(q)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&id)
	}
	return id
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

func logger(file string) {
	f, err := os.OpenFile(doneLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	if _, err := f.WriteString(file + "\n"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func markDone(file string) {
	finished[file] = true
	logger(file)
}

func alreadyRun(file string) bool {
	return finished[file]
}
