package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-ical"
)

const (
	Layout = "2006-01-02 15:04:05 UTC"
	TZ     = "Europe/Lisbon"
)

type EventFormatter func(source EventSource, summary string, dateTime string) string

type EventSource struct {
	Name         string         `json:"name"`
	URL          string         `json:"url"`
	Channel      string         `json:"channel"`
	Filter       string         `json:"filter"`
	Tags         string         `json:"tags"`
	Formatter    EventFormatter `json:"-"`
	FormatterKey string         `json:"formatter"`
}

func removeNonASCII(input string) string {
	output := make([]rune, 0, len(input))

	for _, r := range input {
		if r >= 0 && r <= 127 {
			output = append(output, r)
		}
	}

	return string(output)
}

func getDecoder(url string) (*ical.Decoder, *http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}

	return ical.NewDecoder(resp.Body), resp, nil
}

func getSources() ([]EventSource, error) {
	f, err := os.Open("sources.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var sources []EventSource
	dec := json.NewDecoder(f)
	if err := dec.Decode(&sources); err != nil {
		return nil, err
	}

	for i, source := range sources {
		switch source.FormatterKey {
		case "f1Formatter":
			sources[i].Formatter = f1Formatter
		case "f2Formatter":
			sources[i].Formatter = f2Formatter
		case "f3Formatter":
			sources[i].Formatter = f3Formatter
		case "indyCarFormatter":
			sources[i].Formatter = indyCarFormatter
		case "motoGPFormatter":
			sources[i].Formatter = motoGPFormatter
		case "spaceFormatter":
			sources[i].Formatter = spaceFormatter
		default:
			return nil, errors.New("Unknown formatter")
		}
	}

	return sources, nil

}

func printEvents(dec *ical.Decoder, source EventSource, formatter EventFormatter, ch chan string) {
	for {
		cal, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		for _, event := range cal.Events() {
			location, err := time.LoadLocation(TZ)
			if err != nil {
				log.Fatal(err)
			}

			dt, err := event.DateTimeStart(location)
			if err != nil {
				log.Fatal(err)
			}

			if dt.Before(time.Now()) {
				continue
			}

			summary, err := event.Props.Text(ical.PropSummary)
			if err != nil {
				log.Fatal(err)
			}

			matches := 0
			for _, word := range strings.Split(source.Filter, " ") {
				if strings.Contains(strings.ToLower(summary), strings.ToLower(word)) {
					matches += 1
				}
			}

			if source.Filter == "" || matches > 0 {
				ch <- removeNonASCII(formatter(source, summary, dt.Format(Layout)))
			}
		}
	}
}

func main() {
	var wg sync.WaitGroup
	ch := make(chan string)

	sources, err := getSources()
	if err != nil {
		log.Fatal(err)
	}

	for _, source := range sources {
		wg.Add(1)
		go func() {
			defer wg.Done()

			dec, resp, err := getDecoder(source.URL)
			if err != nil {
				log.Printf("Could not fetch %v calendar\n", source.Name)

				return
			}
			defer resp.Body.Close()

			printEvents(dec, source, source.Formatter, ch)
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for line := range ch {
		fmt.Println(line)
	}
}
