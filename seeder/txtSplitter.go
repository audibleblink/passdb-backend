package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/korovkin/limiter"
)

var (
	mailRE    = regexp.MustCompile(`^([a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+)(.*)$`)
	userLog   = "usernames.txt"
	domainLog = "domains.txt"
	passwdLog = "passwords.txt"
	errLog    = "error.log"

	//procs - 1
	routines = 7
)

func main() {
	f, err := os.OpenFile(errLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	var txtFile string
	txtFile = os.Args[1]

	txtFileHandle, err := os.Open(txtFile)
	if err != nil {
		log.Fatal(err)
	}
	defer txtFileHandle.Close()

	fmt.Println("Starting: " + txtFile)
	for {
		var wg sync.WaitGroup
		lineCh := make(chan string, routines)

		// kick of job listener...
		limit := limiter.NewConcurrencyLimiter(routines)
		go func(
			wgi *sync.WaitGroup,
			ch chan string,
			lim *limiter.ConcurrencyLimiter,
		) {
			for line := range ch {
				lim.Execute(func() {
					processAndSave(wgi, line)
				})
			}
		}(&wg, lineCh, limit)

		// iterate through the lines in the file, adding each to the workgroup
		// before dispatching the line to the processing listener
		lineReader := bufio.NewScanner(txtFileHandle)
		for lineReader.Scan() {
			wg.Add(1)
			lineCh <- lineReader.Text()
		}

		wg.Wait()
	}
}

func processAndSave(wg *sync.WaitGroup, lineText string) {
	defer wg.Done()
	user, domain, password := parse(lineText)
	upsert(user, domain, password)
}

func upsert(user, domain, password string) (err error) {
	logger(userLog, strings.ToLower(user))
	logger(domainLog, strings.ToLower(domain))
	logger(passwdLog, password)
	return
}

func parse(line string) (user, domain, password string) {
	user, domain, password = "nil", "nil", "nil"

	matches := mailRE.FindSubmatch([]byte(line))

	fmt.Println(matches)
	if len(matches) != 3 {
		return
	}

	email := string(matches[1])

	userAndDom := strings.Split(email, "@")
	user = userAndDom[0]
	domain = userAndDom[1]

	passwdWithSeperator := matches[2]

	if len(passwdWithSeperator) > 1 {
		password = string(passwdWithSeperator[1:])
	}
	return
}

// commits text to a file. primarily used to append filenames that
// have been processed already
func logger(file, line string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := f.WriteString(line + "\n"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
