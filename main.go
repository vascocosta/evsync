package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-ical"
)

const ConfigFile = ".evsync.json"

type EventFormatter func(source EventSource, summary string, dateTime string) string

type Config struct {
	Layout  string        `json:"layout"`
	TZ      string        `json:"tz"`
	Sources []EventSource `json:"sources"`
}

type EventSource struct {
	Name         string         `json:"name"`
	URL          string         `json:"url"`
	Channel      string         `json:"channel"`
	Filter       string         `json:"filter"`
	Tags         string         `json:"tags"`
	Notify       bool           `json:"notify"`
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

func getConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ConfigFile)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config Config
	dec := json.NewDecoder(f)
	if err := dec.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func getDecoder(url string) (*ical.Decoder, *http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}

	return ical.NewDecoder(resp.Body), resp, nil
}

func getSources(config *Config) ([]EventSource, error) {
	for i, source := range config.Sources {
		switch source.FormatterKey {
		case "f1Formatter":
			config.Sources[i].Formatter = f1Formatter
		case "f2Formatter":
			config.Sources[i].Formatter = f2Formatter
		case "f3Formatter":
			config.Sources[i].Formatter = f3Formatter
		case "indyCarFormatter":
			config.Sources[i].Formatter = indyCarFormatter
		case "motoGPFormatter":
			config.Sources[i].Formatter = motoGPFormatter
		case "spaceFormatter":
			config.Sources[i].Formatter = spaceFormatter
		case "defaultFormatter":
			config.Sources[i].Formatter = defaultFormatter
		default:
			return nil, errors.New("Unknown formatter")
		}
	}

	return config.Sources, nil

}

func printEvents(config *Config, dec *ical.Decoder, source EventSource, formatter EventFormatter, ch chan string) {
	for {
		cal, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(err)
			continue
		}

		for _, event := range cal.Events() {
			location, err := time.LoadLocation(config.TZ)
			if err != nil {
				log.Fatal(err)
			}

			dt, err := event.DateTimeStart(location)
			if err != nil {
				log.Println(err)
				continue
			}

			if dt.Before(time.Now()) {
				continue
			}

			summary, err := event.Props.Text(ical.PropSummary)
			if err != nil {
				log.Println(err)
				continue
			}

			if strings.Contains(summary, ",") {
				summary = strings.ReplaceAll(summary, ",", "")
			}

			matches := 0
			for _, word := range strings.Split(source.Filter, " ") {
				if strings.Contains(strings.ToLower(summary), strings.ToLower(word)) {
					matches += 1
				}
			}

			if source.Filter == "" || matches > 0 {
				ch <- removeNonASCII(formatter(source, summary, dt.Format(config.Layout)))
			}
		}
	}
}

func main() {
	var wg sync.WaitGroup
	ch := make(chan string)

	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}
	sources, err := getSources(config)
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

			printEvents(config, dec, source, source.Formatter, ch)
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
