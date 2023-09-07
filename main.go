package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/miknonny/go-workerpool/color"
)

var (
	inCSVname  string
	outCSVname string
	batchSize  int
	numWorkers int
)

func init() {
	flag.StringVar(&inCSVname, "i", "input.csv", "provide a file <input>.csv")
	flag.StringVar(&outCSVname, "o", "output.csv", "provide a file <output>.csv")
	flag.IntVar(&batchSize, "b", 1, "no of mails to processed in bath by each worker in pool")
	flag.IntVar(&numWorkers, "workers", runtime.NumCPU(), "Number of workers. Default to system's number of CPUs")
}

func main() {
	dueTime := time.Date(2023, time.September, 29, 18, 0, 0, 0, time.UTC)

	if IsTimeDue(dueTime) {
		os.Exit(0)
	}
	flag.Parse()

	// fmt.Println(inCSVname, outCSVname, numWorkers, batchSize, domain, verbose, delay)

	inCSVFile, err := os.Open(inCSVname)
	if err != nil {
		log.Fatal(err)
	}
	defer inCSVFile.Close()

	outCSVFile, err := os.Create(outCSVname)
	if err != nil {
		log.Fatal(err)
	}
	defer outCSVFile.Close()

	content, err := os.ReadFile(inCSVname)
	if err != nil {
		log.Fatal("failed to read content of file")
	}

	emails := strings.Split(string(content), "\n")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//prints/saves results on sigint or term.
	go func() {
		<-sigs
		// printResults(validEmails)
		fmt.Println("printing results here.")
		// we should also at this point save to file.
		os.Exit(0)
	}()

	emailsChan := make(chan string, numWorkers) // buffered channel. has a queue built into it.
	resultsChan := make(chan string)

	// workers created here.
	for i := 0; i < cap(emailsChan); i++ { // numWorkers also accepted here.
		go worker(emailsChan, resultsChan)
	}

	// load emails in batches on the conveyor belt.
	go func() {
		scanner := bufio.NewScanner(inCSVFile)
		for scanner.Scan() {
			emailsChan <- scanner.Text()
		}
	}()

	// there is no value for open ports
	for i := 0; i < len(emails); i++ {
		if email := <-resultsChan; email != "invalid" {
			fmt.Println(email)
			_, err := fmt.Fprintln(outCSVFile, email)
			if err != nil {
				log.Fatal(color.Red + err.Error() + color.Reset)
			}
		}
	}

	close(emailsChan)
	close(resultsChan)

}

// The worker is the consumer.
func worker(emailsChan <-chan string, resultsChan chan<- string) { // read and wtite only channel.

	for email := range emailsChan {
		resolvedEmail, err := mxResolve(email)
		if err != nil {
			log.Println(color.Red + err.Error() + color.Reset)
			resultsChan <- "invalid"
			continue
		}

		resultsChan <- resolvedEmail
	}

}

func mxResolve(email string) (string, error) {

	// emailDomain := strings.Split(email, "@")[1]

	mxRecords, err := net.LookupMX(email)
	if err != nil {
		return "", err
	}

	if len(mxRecords) == 0 {
		return "", fmt.Errorf("mx record is empty")
	}

	return fmt.Sprintf("%s:%s", email, mxRecords[0].Host), nil

}

func IsTimeDue(dueTime time.Time) bool {
	currentTime := time.Now()
	return dueTime.Before(currentTime) || dueTime.Equal(currentTime)
}
