package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const dateFormat = "2006-01-02"

func main() {
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal(err)
	}
	start := flag.String("start", "2014-01-01", "start date YYYY-MM-DD")
	end := flag.String("end", time.Now().In(est).Format(dateFormat), "end date YYYY-MM-DD")
	channel := flag.String("chan", "", "channel")
	workers := flag.Int("workers", 4, "number of workers")
	flag.Parse()

	startTime, err := time.ParseInLocation(dateFormat, *start, est)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = time.ParseInLocation(dateFormat, *end, est); err != nil {
		log.Fatal(err)
	}
	if *channel == "" {
		log.Fatal("missing channel")
	}

	var wg sync.WaitGroup
	dates := make(chan string)
	wg.Add(*workers)
	for i := 0; i < *workers; i++ {
		go downloadLogs(*channel, dates, &wg)
	}

	for t := startTime; ; t = t.Add(24 * time.Hour) {
		date := t.Format(dateFormat)
		dates <- date
		if date == *end {
			break
		}
	}

	close(dates)
	wg.Wait()
}

func downloadLogs(channel string, dates <-chan string, done *sync.WaitGroup) {
	for date := range dates {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://irclogger.com/.%s/%s", channel, date), nil)
		req.Header.Set("Accept", "text/plain")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		out, err := os.Create(date + ".txt")
		if err != nil {
			log.Fatal(err)
		}

		if _, err := io.Copy(out, res.Body); err != nil {
			log.Fatal(err)
		}
		res.Body.Close()
		out.Close()
		fmt.Println(date)
	}
	done.Done()
}
