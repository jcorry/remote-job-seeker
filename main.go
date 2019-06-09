package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type XMLRss struct {
	XMLName xml.Name `xml:"rss"`
	Channel Jobs
}

type Jobs struct {
	XMLName  xml.Name `xml:"channel", json:"-"`
	JobsList []*Job   `xml:"item"`
}

type Job struct {
	ID          string    `json:"id" xml:"guid,string"`
	DateString  string    `json:"created_at" xml:"pubDate,string"`
	Date        time.Time `json:"created_datetime"`
	Company     string    `json:"company"`
	CompanyURL  string    `json:"company_url"`
	Location    string    `json:"location"`
	Position    string    `json:"position" xml:"title,string"`
	Apply       string    `json:"how_to_apply" xml:"link,string"`
	Source      string    `json:"source"`
	Description string    `json:"description" xml:"description,string"`
}

type Formatter interface {
	Format(res []byte) []*Job
}

type Source struct {
	URL string
	Formatter
}

func main() {
	sources := []Source{
		{
			URL:       "https://jobs.github.com/positions.json?&location=remote",
			Formatter: new(GitHubFormatter),
		},
		{
			URL:       "https://stackoverflow.com/jobs/feed?r=true",
			Formatter: new(StackOverflowFormatter),
		},
		{
			URL:       "https://remoteok.io/remote-jobs.rss",
			Formatter: new(RemoteOkFormatter),
		},
	}

	var wg sync.WaitGroup

	var fileData []*Job

	start := time.Now()

	wg.Add(len(sources))
	for _, s := range sources {
		url := s.URL
		formatter := s
		go func() {
			defer wg.Done()
			fmt.Println(fmt.Sprintf("Fetching URL: %s", url))
			bytes, err := fetch(url)
			if err != nil {
				fmt.Errorf("%s", err)
			}

			fileData = append(fileData, formatter.Format(bytes)...)
		}()
	}
	wg.Wait()

	writeJson(fileData)
	fmt.Println(fmt.Sprintf("Captured %d records", len(fileData)))
	fmt.Println(fmt.Sprintf("Ran in: %s", time.Since(start)))
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var data bytes.Buffer
	_, err = io.Copy(&data, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func writeJson(jobs []*Job) {
	f, err := os.Create("jobs.json")
	if err != nil {
		fmt.Println(fmt.Errorf("%s", err))
	}
	defer f.Close()

	e := json.NewEncoder(f)
	e.SetEscapeHTML(false)

	err = e.Encode(jobs)
	if err != nil {
		fmt.Println(fmt.Errorf("%s", err))
	}
}

// Formatters for each source

// GitHubFormatter formats the JSON returned from GitHub Jobs
type GitHubFormatter struct{}

func (g *GitHubFormatter) Format(res []byte) []*Job {
	var jobs []*Job
	err := json.NewDecoder(bytes.NewReader(res)).Decode(&jobs)
	if err != nil {
		log.Fatal(err)
	}

	for _, j := range jobs {
		t, _ := time.Parse("Mon Jan _2 15:04:05 UTC 2006", j.DateString)
		j.Date = t
		j.Source = "Github Jobs"
	}

	return jobs
}

// StackOverflowFormatter formats the JSON returned from Stack Overflow Jobs
type StackOverflowFormatter struct{}

func (s *StackOverflowFormatter) Format(res []byte) []*Job {
	var rss XMLRss
	err := xml.NewDecoder(bytes.NewReader(res)).Decode(&rss)
	if err != nil {
		log.Fatal(err)
	}
	jobs := rss.Channel.JobsList
	for _, j := range jobs {
		t, _ := time.Parse("Mon, _2 Jan 2006 15:04:05 Z", j.DateString)
		j.Date = t
		j.Source = "Stack Overflow"
	}
	return jobs
}

type RemoteOkFormatter struct{}

func (r *RemoteOkFormatter) Format(res []byte) []*Job {
	type job struct {
		ID          string    `xml:"guid"`
		DateString  string    `xml:"pubDate"`
		Date        time.Time `xml:"-"`
		Position    string    `xml:"title"`
		Company     string    `xml:"company"`
		Source      string    `xml:"-"`
		Apply       string    `xml:"link"`
		Description string    `xml:"description"`
	}

	type channel struct {
		XMLName  xml.Name `xml:"channel", json:"-"`
		JobsList []*job   `xml:"item"`
	}

	type xmlrss struct {
		XMLName xml.Name `xml:"rss"`
		Channel channel
	}

	var rss xmlrss

	err := xml.NewDecoder(bytes.NewReader(res)).Decode(&rss)
	if err != nil {
		log.Fatal(err)
	}

	jobs := rss.Channel.JobsList

	var data []*Job
	for _, j := range jobs {
		t, _ := time.Parse("2006-01-02T15:04:05-07:00", j.DateString)
		j.Date = t
		j.Source = "RemoteOK Jobs"

		job := &Job{
			ID:          strings.TrimSpace(j.ID),
			DateString:  j.DateString,
			Date:        j.Date,
			Position:    strings.TrimSpace(j.Position),
			Company:     strings.TrimSpace(j.Company),
			Source:      strings.TrimSpace(j.Source),
			Apply:       strings.TrimSpace(j.Apply),
			Description: j.Description,
		}

		data = append(data, job)
	}

	return data
}
